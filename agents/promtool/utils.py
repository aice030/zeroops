#

# utils for our web-agent

import re
import os
import subprocess
import signal
import time
import requests
import base64
import json
from ..base.utils import KwargsInitializable, rprint, zwarn, zlog
import pandas as pd
import numpy as np
from statsmodels.tsa.seasonal import STL
from datetime import datetime, timedelta
import json
import warnings

# --
# web state
class PromState:
    def __init__(self, **kwargs):
        # not-changed
        self.browser_id = ""
        self.page_id = ""
        self.target_url = ""
        self.target_prometheus_metrics = ""  # Add this for Prometheus
        
        # step info
        self.curr_step = 0  # step to the root
        self.curr_screenshot_mode = False  # whether we are using screenshot or not?
        self.total_actual_step = 0  # [no-rev] total actual steps including reverting (can serve as ID)
        self.num_revert_state = 0  # [no-rev] number of state reversion
        # (last) action information
        self.action_string = ""
        self.action = None
        self.error_message = ""
        # --
        self.update(**kwargs)

    def get_id(self):  # use these as ID
        return (self.browser_id, self.page_id, self.total_actual_step)

    def update(self, **kwargs):
        for k, v in kwargs.items():
            assert (k in self.__dict__), f"Attribute not found for {k} <- {v}"
        self.__dict__.update(**kwargs)

    def to_dict(self):
        return self.__dict__.copy()

    def copy(self):
        return PromState(**self.to_dict())

    def __repr__(self):
        return f"PromState({self.__dict__})"



# an opened web browser
class PromEnv(KwargsInitializable):
    def __init__(self, starting=True, starting_target_url=None, **kwargs):
        self.prom_ip = os.getenv("PROM_IP", "localhost:9090")  # allow set by ENV
        self.prom_timeout = 600  # set a timeout!
        self.target_url = starting_target_url if starting_target_url else "http://localhost:9090"  # default Prometheus URL
        # --
        super().__init__(**kwargs)
        # --
        self.state: PromState = None
        self.prom_metrics_path = ""
        # --
        if starting:
            self.start(starting_target_url)  # start at the beginning
        # --

    def start(self, target_url=None):
        self.stop()  # stop first
        # --
        target_url = target_url if target_url is not None else self.target_url  # otherwise use default
        self.init_state(target_url)

    def stop(self):
        # For Prometheus, we don't need to close browser
        # Just clean up the state
        self.state = None

    def __del__(self):
        self.stop()

    # note: return a copy!
    def get_state(self, export_to_dict=True, return_copy=True):
        assert self.state is not None, "Current state is None, should first start it!"
        if export_to_dict:
            ret = self.state.to_dict()
        elif return_copy:
            ret = self.state.copy()
        else:
            ret = self.state
        return ret

    def get_status(self):
        """Get Prometheus connection status and basic info"""
        try:
            # Test connection to Prometheus
            response = requests.get(f"http://{self.prom_ip}/api/v1/status/config", timeout=10)
            if response.status_code == 200:
                return {
                    "status": "connected",
                    "available_metrics": self._get_available_metrics(),
                    "query_history": getattr(self.state, 'query_history', []) if self.state else [],
                    "last_query_result": getattr(self.state, 'last_query_result', "N/A") if self.state else "N/A"
                }
            else:
                return {
                    "status": f"error: HTTP {response.status_code}",
                    "available_metrics": [],
                    "query_history": [],
                    "last_query_result": "N/A"
                }
        except Exception as e:
            return {
                "status": f"error: {e}",
                "available_metrics": [],
                "query_history": [],
                "last_query_result": "N/A"
            }

    def _get_available_metrics(self):
        """Get list of available metrics from Prometheus"""
        try:
            response = requests.get(f"http://{self.prom_ip}/api/v1/label/__name__/values", timeout=10)
            if response.status_code == 200:
                data = response.json()
                return data.get('data', [])[:50]  # Limit to first 50 metrics
            else:
                return []
        except Exception as e:
            zwarn(f"Failed to get available metrics: {e}")
            return []

    def get_prom_data(self):
        return self.prom_metrics_path
    
    def prometheus_analysis(self, metrics_path):
        return self.prom_analysis_result

    def get_target_url(self):
        return self.target_url

    def step_state(self, action_string: str):
        """Execute a Prometheus action (placeholder for compatibility)"""
        return f"Executed action: {action_string}"

    def sync_files(self):
        """Sync any downloaded files (placeholder for Prometheus)"""
        # For Prometheus, this might not be needed, but keeping for compatibility
        pass

    def init_state(self, target_url: str):
        # For Prometheus, we don't need browser_id and page_id
        # Just initialize the state with basic information
        curr_step = 0
        state = PromState(
            browser_id="", 
            page_id="", 
            target_url=target_url, 
            curr_step=curr_step, 
            total_actual_step=curr_step,
            target_prometheus_metrics=target_url
        )  # start from 0
        # --
        self.state = state  # set the new state!
        # --

    def end_state(self):
        # For Prometheus, we don't need to close browser
        # Just clean up the state
        self.state = None

class AnomalyDetector:
    def __init__(self, file_path='yzh_mirror_data.csv',lower_percentile=0.5, upper_percentile=99.5):
        """
        初始化异常检测器
        
        Args:
            data_file: 数据文件路径
            lower_percentile (float): 下界百分位数，默认0.5%
            upper_percentile (float): 上界百分位数，默认99.5%
        """
        self.file_path = file_path
        self.lower_percentile = lower_percentile
        self.upper_percentile = upper_percentile
    
    def load_data(self, file_path=None):
        """
        加载CSV数据文件
        
        Args:
            file_path (str): CSV文件路径，如果为None则使用self.file_path
            
        Returns:
            pandas.DataFrame: 加载的数据
        """
        try:
            if file_path is None:
                file_path = self.file_path
            df = pd.read_csv(file_path)
            # 确保时间列格式正确
            df['time'] = pd.to_datetime(df['time'])
            return df
        except Exception as e:
            raise Exception(f"数据加载失败: {str(e)}")
    
    def stl_decomposition(self, data, period=None):
        """
        执行STL分解
        
        Args:
            data (array-like): 时间序列数据
            period (int): 季节性周期，如果为None则自动推断
            
        Returns:
            dict: 包含trend, seasonal, residual的字典
        """
        if period is None:
            # 自动推断周期，这里假设数据是3秒间隔，一天有28800个点
            # 可以根据实际数据调整
            period = min(len(data) // 4, 28800)  # 避免周期过大
        
        # 确保周期至少为2
        period = max(2, period)
        
        try:
            stl_result = STL(data, period=period, robust=True).fit()
            return {
                'trend': stl_result.trend,
                'seasonal': stl_result.seasonal,
                'residual': stl_result.resid
            }
        except Exception as e:
            # 如果STL分解失败，使用简单的移动平均作为趋势
            print(f"STL分解失败，使用移动平均替代: {str(e)}")
            trend = pd.Series(data).rolling(window=min(10, len(data)//10), center=True).mean()
            trend = trend.fillna(method='bfill').fillna(method='ffill')
            residual = data - trend
            return {
                'trend': trend,
                'seasonal': np.zeros_like(data),
                'residual': residual
            }
    
    def detect_anomalies(self, data, metric_name, host):
        """
        检测异常值
        
        Args:
            data (pandas.DataFrame): 包含time和value列的数据
            metric_name (str): 指标名称
            host (str): 主机地址
            
        Returns:
            list: 异常值列表
        """
        # 按时间排序
        data = data.sort_values('time').reset_index(drop=True)
        
        # 执行STL分解
        decomposition = self.stl_decomposition(data['value'].values)
        residuals = decomposition['residual']
        
        # 计算残差的百分位数
        lower_bound = np.percentile(residuals, self.lower_percentile)
        upper_bound = np.percentile(residuals, self.upper_percentile)
        
        # 找出异常值
        anomalies = []
        for idx, (time, value, residual) in enumerate(zip(data['time'], data['value'], residuals)):
            if residual < lower_bound or residual > upper_bound:
                # 判断异常类型
                if residual < lower_bound:
                    anomaly_type = "偏低"
                    threshold = lower_bound
                else:
                    anomaly_type = "偏高"
                    threshold = upper_bound
                
                # 获取单位（从指标名称推断）
                unit = self._extract_unit(metric_name)
                
                anomalies.append({
                    "label": f"{metric_name}{{host=\"{host}\"}}",  # PromQL格式
                    "host": host,
                    "startTime": time.strftime("%Y-%m-%dT%H:%M:%S+00:00"),
                    "endTime": time.strftime("%Y-%m-%dT%H:%M:%S+00:00"),
                    "异常描述": f"({metric_name},{unit},{anomaly_type})",
                    "original_value": float(value),
                    "residual": float(residual),
                    "threshold": float(threshold),
                    "anomaly_type": anomaly_type
                })
        
        return anomalies
    
    def _extract_unit(self, metric_name):
        """
        从指标名称中提取单位
        
        Args:
            metric_name (str): 指标名称
            
        Returns:
            str: 单位
        """
        metric_name_lower = metric_name.lower()
        
        if 'cpu' in metric_name_lower and 'percentage' in metric_name_lower:
            return "百分比"
        elif 'memory' in metric_name_lower:
            return "字节"
        elif 'disk' in metric_name_lower:
            return "字节"
        elif 'network' in metric_name_lower:
            return "字节/秒"
        elif 'response_time' in metric_name_lower or 'latency' in metric_name_lower:
            return "毫秒"
        else:
            return "单位"
    
    def process_file(self, file_path):
        """
        处理整个文件并检测异常
        
        Args:
            file_path (str): CSV文件路径
            
        Returns:
            dict: 包含异常检测结果的字典
        """
        # 加载数据
        df = self.load_data(file_path)
        
        # 按指标名称和主机分组
        results = []
        
        for (metric_name, host), group in df.groupby(['name', 'host']):
            anomalies = self.detect_anomalies(group, metric_name, host)
            results.extend(anomalies)
        
        return {
            "total_anomalies": len(results),
            "anomalies": results,
            "detection_params": {
                "lower_percentile": self.lower_percentile,
                "upper_percentile": self.upper_percentile,
                "method": "STL分解 + 百分位数检测"
            }
        }

