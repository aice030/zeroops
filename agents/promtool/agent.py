#

import os
import re
import shutil
import urllib.request
import json
from contextlib import contextmanager
import time
import pandas as pd

from ..base.agent import MultiStepAgent, register_template, ActionResult
from ..base.model import LLM
from ..base.utils import zwarn, rprint, have_images_in_messages
from ..base.tool import SimpleSearchTool
from .utils import AnomalyDetector

from .utils import PromEnv
from .prompts import PROMPTS as PROM_PROMPTS


class Prom_ToolAgent(MultiStepAgent):
    def __init__(self, **kwargs):
        # note: this is a little tricky since things will get re-init again in super().__init__
        feed_kwargs = dict(
            name="prom_agent",
            description="A Prometheus agent helping to get and analyze prometheus metrics and return the results.",
            templates={"plan": "prom_plan", "action": "prom_action", "end": "prom_end"},  # template names
            max_steps=5,
        )
        feed_kwargs.update(kwargs)
        self.prom_env_kwargs = {}  # kwargs for prometheus env
        self.use_multimodal = "auto"  # no: always no, yes: always yes, auto: let the agent decide
        # --
        register_template(PROM_PROMPTS)  # add web prompts
        super().__init__(**feed_kwargs)
        # 重新设置model为fake模式，避免用户输入
        self.model = LLM(call_target="gpt:gpt-oss-20b")  # llm model for testing
        self.prom_envs = {}  # session_id -> ENV
        # Define Prometheus-specific functions
        self.ACTIVE_FUNCTIONS.update(
            fetch_and_analyze_prometheus_data=self._fetch_and_analyze_prometheus_data,  # 抓取并分析Prometheus数据
            stop=self._my_stop,
            save=self._my_save
        )
        # --

    # Prometheus data functions - 合并版本
    def _fetch_and_analyze_prometheus_data(self, query: str, start_time: str = None, end_time: str = None, 
                                          step: str = None, analysis_type: str = "general", 
                                          return_data: bool = True):
        """
        抓取并分析Prometheus数据（合并版本）
        Args:
            query: Prometheus查询语句
            start_time: 开始时间
            end_time: 结束时间
            step: 步长
            analysis_type: 分析类型 ("general", "trend_analysis", "anomaly_detection")
            return_data: 是否返回原始数据
        Returns:
            ActionResult: 包含抓取结果、分析结果和自然语言解读的完整结果
        """
        try:
            # 步骤1: 抓取Prometheus数据
            print(f"🔍 开始抓取Prometheus数据: {query}")
            
            # 模拟抓取过程 - 这里应该调用实际的Prometheus查询函数
            csv_file_path = "agents/promtool/tmp/heap_memory_filtered.csv"
            df = pd.read_csv(csv_file_path)
            
            fetched_data = {
                "query": query,
                "start_time": start_time,
                "end_time": end_time,
                "step": step,
                "data_points": df.to_dict('records'),
                "metadata": {
                    "total_points": len(df),
                    "query_duration": "1.2s",
                    "status": "success"
                }
            }
            
            fetch_result = f"Successfully fetched Prometheus data for query: {query}"
            if start_time and end_time:
                fetch_result += f" from {start_time} to {end_time}"
            if step:
                fetch_result += f" with step {step}"
            fetch_result += f". Data points: {len(fetched_data['data_points'])}"
            
            print(f"✅ 数据抓取完成: {len(fetched_data['data_points'])} 个数据点")
            
            # 步骤2: 分析数据
            print(f"📈 开始分析数据，分析类型: {analysis_type}")
            
            # 执行分析
            if analysis_type == "trend_analysis":
                values = [point["value"] for point in fetched_data.get("data_points", [])]
                if values:
                    avg_value = sum(values) / len(values)
                    trend = "上升" if values[-1] > values[0] else "下降" if values[-1] < values[0] else "稳定"
                    analysis_result = f"平均值: {avg_value:.2f}, 趋势: {trend}"
                else:
                    analysis_result = "无数据点可分析"
            elif analysis_type == "anomaly_detection":
                # 调用异常检测API
                try:
                    detector = AnomalyDetector('agents/promtool/tmp/heap_memory_filtered.csv')
                    report = detector.process_file('agents/promtool/tmp/heap_memory_filtered.csv')
                    analysis_result = f"异常检测报告: {report}"
                except Exception as e:
                    analysis_result = f"异常检测失败: {e}"
            else:
                analysis_result = f"通用分析完成，数据点数量: {len(fetched_data.get('data_points', []))}"

            print(f"✅ 数据分析完成: {analysis_result}")

            # 步骤3: 使用LLM解读分析结果
            print("🤖 开始LLM解读分析结果")
            natural_language_result = self._interpret_analysis_with_llm(
                analysis_result, 
                fetched_data, 
                analysis_type
            )
            print(f"✅ LLM解读完成")

            # 步骤4: 构建完整结果
            complete_result = f"=== Prometheus数据抓取与分析报告 ===\n\n"
            complete_result += f"📊 数据抓取:\n{fetch_result}\n\n"
            complete_result += f"📈 数据分析:\n{analysis_result}\n\n"
            complete_result += f"🤖 自然语言解读:\n{natural_language_result}"
            
            # 保存最后抓取的数据（向后兼容）
            self._last_fetched_data = fetched_data
            
            # 返回完整结果
            return ActionResult(
                "fetch_and_analyze_prometheus_data", 
                complete_result,
                data=fetched_data if return_data else None,
                analysis_result=analysis_result,
                natural_language_result=natural_language_result,
                fetch_result=fetch_result
            )
            
        except Exception as e:
            error_msg = f"Failed to fetch and analyze Prometheus data: {e}"
            print(f"❌ 错误: {error_msg}")
            return ActionResult("fetch_and_analyze_prometheus_data", error_msg)



    def _interpret_analysis_with_llm(self, analysis_result, data, analysis_type):
        """
        使用LLM解读分析结果并转换为自然语言
        
        Args:
            analysis_result: 原始分析结果
            data: 原始数据
            analysis_type: 分析类型
            
        Returns:
            str: 自然语言解读结果
        """
        try:
            # 构建提示词
            prompt = f"""
                你是一个专业的Prometheus监控数据分析专家。请将以下分析结果转换为清晰、易懂的自然语言描述。

                分析类型: {analysis_type}
                原始数据: {json.dumps(data, ensure_ascii=False, indent=2)}
                分析结果: {analysis_result}

                请提供：
                1. 数据概览的简要描述
                2. 关键发现和洞察
                3. 如果有异常或趋势，请详细说明
                4. 对运维或业务的影响分析
                5. 建议的后续行动

                请用中文回答，语言要专业但易懂，适合技术团队阅读。
                """
                        
            # 调用LLM进行解读
            messages = [
                {"role": "system", "content": "你是一个专业的Prometheus监控数据分析专家，擅长将技术分析结果转换为清晰易懂的自然语言描述。"},
                {"role": "user", "content": prompt}
            ]
            
            # 使用实例的LLM模型
            llm_response = self.model(messages)
            
            # 提取LLM的回复内容
            if hasattr(llm_response, 'content'):
                natural_language = llm_response.content
            elif isinstance(llm_response, str):
                natural_language = llm_response
            else:
                natural_language = str(llm_response)
            
            return natural_language
            
        except Exception as e:
            # 如果LLM解读失败，返回原始结果
            zwarn(f"LLM interpretation failed: {e}")
            return f"分析结果解读失败，原始结果: {analysis_result}"

    # note: a specific stop function!
    def _my_stop(self, answer: str = None, summary: str = None, output: str = None):
        if output:
            ret = f"Final answer: [{output}] ({summary})"
        else:
            ret = f"Final answer: [{answer}] ({summary})"
        self.put_final_result(ret)  # mark end and put final result
        return ActionResult("stop", ret)

    # note: special save
    def _my_save(self, remote_path: str, local_path: str):
        try:
            _dir = os.path.dirname(local_path)
            if _dir:
                os.makedirs(_dir, exist_ok=True)
            if local_path != remote_path:
                remote_path = remote_path.strip()
                if remote_path.startswith("http://") or remote_path.startswith("https://"):  # retrieve from the web
                    urllib.request.urlretrieve(remote_path, local_path)
                else:  # simply copy!
                    shutil.copyfile(remote_path, local_path)
            ret = f"Save Succeed: from remote_path = {remote_path} to local_path = {local_path}"
        except Exception as e:
            ret = f"Save Failed with {e}: from remote_path = {remote_path} to local_path = {local_path}"
        return ActionResult("save", ret)


    def get_function_definition(self, short: bool):
        if short:
            return "- def prom_agent(task: str) -> Dict:  # Fetches Prometheus metrics data and analyzes it to return results."
        else:
            return """- prom_agent
```python
def prom_agent(task: str) -> dict:
    \""" Fetches fetch real‑time Prometheus metrics data and analyzes it to return results.
    Args:
        task (str): A detailed description of the task to perform. This may include:
            - The specific Prometheus metrics to fetch (query, time range, etc.)
            - The type of analysis required
            - Specific output format requirements
    Returns:
        dict: A dictionary with the following structure:
            {
                'output': <str>  # The well-formatted answer, strictly following any specified output format.
                'log': <str>     # Additional notes, such as steps taken, issues encountered, or relevant context.
            }
    Notes:
        - The agent will first fetch Prometheus data and store it in memory
        - Then it will analyze the data directly without saving to files
        - Data can be passed directly between functions for better performance
        - If the `task` specifies an output format, ensure the 'output' field matches it exactly
    Example:
        >>> answer = prom_agent(task="Fetch CPU usage metrics for the last hour and analyze the trend")
        >>> print(answer)  # directly print the result of the analysis
    \"""
```"""
    #上面的这个Example里面的内容还需要改

    # allow *args styled calling
    def __call__(self, task: str, **kwargs):  # allow *args styled calling
        result = super().__call__(task, **kwargs)
        # Print the result in a format that can be parsed by the main agent
        print(f"PROM_AGENT_RESULT_OUTPUT: {result.output}")
        print(f"PROM_AGENT_RESULT_LOG: {result.log}")
        return result

    def init_run(self, session):
        super().init_run(session)
        _id = session.id
        assert _id not in self.prom_envs
        _kwargs = self.prom_env_kwargs.copy()
        if session.info.get("target_prometheus_metrics"):
            _kwargs["starting_target_prometheus_metrics"] = session.info["target_prometheus_metrics"]
        self.prom_envs[_id] = PromEnv(**_kwargs)

    def end_run(self, session):
        ret = super().end_run(session)
        _id = session.id
        self.prom_envs[_id].stop()
        del self.prom_envs[_id]  # remove prom env
        return ret

    def step_call(self, messages, session, model=None):
        if model is None:
            model = self.model # use which model?
        response = model(messages)
        return response

    def step_prepare(self, session, state):
        _input_kwargs, _extra_kwargs = super().step_prepare(session, state)
        _prom_env = self.prom_envs[session.id]

        _prom_state = _prom_env.get_status()
        _input_kwargs.update({
            "prometheus_status": _prom_state["status"],
            "available_metrics": _prom_state.get("available_metrics", []),
            "query_history": _prom_state.get("query_history", []),
            "last_query_result": _prom_state.get("last_query_result", "N/A")
        })
        
        _extra_kwargs["prom_env"] = _prom_env
        return _input_kwargs, _extra_kwargs

    def step_action(self, action_res, action_input_kwargs, prom_env=None, **kwargs):
        action_res["prom_state_before"] = prom_env.get_status()  # inplace storage of the web-state before the action
        _rr = super().step_action(action_res, action_input_kwargs)  # get action from code execution
        if isinstance(_rr, ActionResult):
            action_str, action_result = _rr.action, _rr.result
        else:
            action_str = self.get_obs_str(None, obs=_rr, add_seq_enum=False)
            action_str, action_result = "nop", action_str.strip()  # no-operation
        # state step
        try:  # execute the action on the prometheus
            step_result = prom_env.step_state(action_str)
            ret = action_result if action_result is not None else step_result  # use action result if there are direct ones
            prom_env.sync_files()
            # ret = f"Browser step: {action_str.strip()}"
        except Exception as e:
            zwarn("prom_env execution error!")
            ret = f"Prometheus error: {e}"
        return ret

    # --
    # other helpers
    def set_multimodal(self, use_multimodal):
        if use_multimodal is not None:
            self.use_multimodal = use_multimodal

    def get_multimodal(self):
        return self.use_multimodal
