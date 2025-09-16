// Mock API服务
import { mockServicesData, mockServiceDetails, mockVersionOptions, mockScheduledReleases, mockServiceActiveVersions, mockServiceMetrics, mockAvailableVersions, mockDeploymentPlans, mockMetricsData, mockDeploymentChangelog, mockAlertRuleChangelog, mockAlertsData, mockAlertDetails, type ServicesResponse, type ServiceDetail, type ServiceActiveVersionsResponse, type ServiceMetricsResponse, type AvailableVersionsResponse, type DeploymentPlansResponse, type MetricsResponse, type DeploymentChangelogResponse, type AlertRuleChangelogResponse, type AlertsResponse, type AlertDetail } from './services'

// 模拟网络延迟
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// Mock API服务类
export class MockApiService {
  // 获取服务列表
  static async getServices(): Promise<ServicesResponse> {
    await delay(500) // 模拟网络延迟
    console.log('Mock API: 获取服务列表')
    return mockServicesData
  }

  // 获取服务详情
  static async getServiceDetail(serviceName: string): Promise<ServiceDetail> {
    await delay(300) // 模拟网络延迟
    console.log(`Mock API: 获取服务详情 - ${serviceName}`)
    
    const serviceDetail = mockServiceDetails[serviceName]
    if (!serviceDetail) {
      throw new Error(`服务 ${serviceName} 不存在`)
    }
    
    return serviceDetail
  }

  // 获取服务活跃版本
  static async getServiceActiveVersions(serviceName: string): Promise<ServiceActiveVersionsResponse> {
    await delay(300)
    console.log(`Mock API: 获取服务活跃版本 - ${serviceName}`)
    const activeVersions = mockServiceActiveVersions[serviceName]
    if (!activeVersions) {
      throw new Error(`服务 ${serviceName} 的活跃版本数据不存在`)
    }
    return activeVersions
  }

  // 获取服务指标数据 - 新的API接口
  static async getServiceMetrics(serviceName: string): Promise<ServiceMetricsResponse> {
    await delay(300) // 模拟网络延迟
    console.log(`Mock API: 获取服务指标数据 - ${serviceName}`)
    
    const metrics = mockServiceMetrics[serviceName]
    if (!metrics) {
      throw new Error(`服务 ${serviceName} 的指标数据不存在`)
    }
    
    return metrics
  }

  // 获取服务可发布版本列表 - 新的API接口
  static async getServiceAvailableVersions(serviceName: string): Promise<AvailableVersionsResponse> {
    await delay(300) // 模拟网络延迟
    console.log(`Mock API: 获取服务可发布版本列表 - ${serviceName}`)
    
    const availableVersions = mockAvailableVersions[serviceName]
    if (!availableVersions) {
      throw new Error(`服务 ${serviceName} 的可发布版本数据不存在`)
    }
    
    return availableVersions
  }

  // 获取服务发布计划列表 - 新的API接口
  static async getServiceDeploymentPlans(serviceName: string): Promise<DeploymentPlansResponse> {
    await delay(300) // 模拟网络延迟
    console.log(`Mock API: 获取服务发布计划列表 - ${serviceName}`)
    
    const deploymentPlans = mockDeploymentPlans[serviceName]
    if (!deploymentPlans) {
      throw new Error(`服务 ${serviceName} 的发布计划数据不存在`)
    }
    
    return deploymentPlans
  }

  // 获取服务指标数据 - 新的API接口
  static async getServiceMetricsData(serviceName: string, metricName: string, version: string): Promise<MetricsResponse> {
    await delay(300) // 模拟网络延迟
    console.log(`Mock API: 获取服务指标数据 - ${serviceName}/${metricName}?version=${version}`)
    
    const serviceMetrics = mockMetricsData[serviceName]
    if (!serviceMetrics) {
      throw new Error(`服务 ${serviceName} 的指标数据不存在`)
    }
    
    const metricData = serviceMetrics[metricName]
    if (!metricData) {
      throw new Error(`服务 ${serviceName} 的指标 ${metricName} 数据不存在`)
    }
    
    return metricData
  }


  // 取消部署计划 - 新的API接口
  static async cancelDeployment(deployID: string): Promise<{ status: number }> {
    await delay(300)
    console.log(`Mock API: 取消部署计划 - ${deployID}`)
    // 模拟删除操作，返回状态码200
    return { status: 200 }
  }

  // 暂停部署计划 - 新的API接口
  static async pauseDeployment(deployID: string): Promise<{ status: number }> {
    await delay(300)
    console.log(`Mock API: 暂停部署计划 - ${deployID}`)
    // 模拟暂停操作，返回状态码200
    return { status: 200 }
  }

  // 继续部署计划 - 新的API接口
  static async continueDeployment(deployID: string): Promise<{ status: number }> {
    await delay(300)
    console.log(`Mock API: 继续部署计划 - ${deployID}`)
    // 模拟继续操作，返回状态码200
    return { status: 200 }
  }

  // 回滚部署计划 - 新的API接口
  static async rollbackDeployment(deployID: string): Promise<{ status: number }> {
    await delay(300)
    console.log(`Mock API: 回滚部署计划 - ${deployID}`)
    // 模拟回滚操作，返回状态码200
    return { status: 200 }
  }

  // 创建部署计划 - 新的API接口
  static async createDeployment(data: {service: string, version: string, scheduleTime?: string}): Promise<{ status: number, data: {id: string, message: string} }> {
    await delay(500)
    console.log(`Mock API: 创建部署计划 - service: ${data.service}, version: ${data.version}`)
    
    // 生成模拟的部署ID
    const deployID = `deploy-${Date.now()}`
    
    // 模拟创建成功，返回状态码201
    return { 
      status: 201,
      data: {
        id: deployID,
        message: 'deployment created successfully'
      }
    }
  }

  // 更新部署计划 - 新的API接口
  static async updateDeployment(deployID: string, data: {version?: string, scheduleTime?: string}): Promise<{ status: number, data: {message: string} }> {
    await delay(300)
    console.log(`Mock API: 更新部署计划 - ${deployID}`, data)
    
    // 模拟更新成功，返回状态码200
    return { 
      status: 200,
      data: {
        message: 'deployment updated successfully'
      }
    }
  }

  // 获取部署变更记录 - 新的API接口
  static async getDeploymentChangelog(start?: string, limit?: number): Promise<DeploymentChangelogResponse> {
    await delay(300)
    console.log(`Mock API: 获取部署变更记录 - start: ${start}, limit: ${limit}`)
    
    // 模拟分页逻辑
    let items = [...mockDeploymentChangelog.items]
    
    // 如果有start参数，模拟从该时间点开始的数据
    if (start) {
      const startTime = new Date(start)
      items = items.filter(item => new Date(item.startTime) <= startTime)
    }
    
    // 如果有limit参数，限制返回数量
    if (limit && limit > 0) {
      items = items.slice(0, limit)
    }
    
    return {
      items,
      next: items.length > 0 ? items[items.length - 1].startTime : undefined
    }
  }

  // 获取告警规则变更记录
  static async getAlertRuleChangelog(start?: string, limit?: number): Promise<AlertRuleChangelogResponse> {
    await delay(400) // 模拟网络延迟
    console.log(`Mock API: 获取告警规则变更记录 - start: ${start}, limit: ${limit}`)

    let items = [...mockAlertRuleChangelog.items]

    // 1. 先按时间排序（从新到旧）
    items.sort((a, b) => new Date(b.editTime).getTime() - new Date(a.editTime).getTime())

    // 2. 根据 start 参数筛选数据（分页逻辑）
    if (start) {
      const startTime = new Date(start)
      items = items.filter(item => new Date(item.editTime) <= startTime)
    }

    // 3. 根据limit限制返回数量
    if (limit && limit > 0) {
      items = items.slice(0, limit)
    }

    return {
      items,
      next: items.length > 0 ? items[items.length - 1].editTime : undefined
    }
  }

  // 获取告警列表
  static async getAlerts(start?: string, limit: number = 10, state?: string): Promise<AlertsResponse> {
    await delay(400) // 模拟网络延迟
    console.log(`Mock API: 获取告警列表 - start: ${start}, limit: ${limit}, state: ${state}`)
    
    let items = [...mockAlertsData.items]
    
    // 1. 先按时间排序（从新到旧）
    items.sort((a, b) => new Date(b.alertSince).getTime() - new Date(a.alertSince).getTime())
    
    // 2. 根据 start 参数筛选数据（分页逻辑）
    if (start) {
      const startTime = new Date(start)
      items = items.filter(alert => new Date(alert.alertSince) <= startTime)
    }
    
    // 3. 根据state参数过滤数据
    if (state) {
      items = items.filter(alert => alert.state === state)
    }
    
    // 4. 根据limit限制返回数量
    if (limit && limit > 0) {
      items = items.slice(0, limit)
    }
    
    return {
      items,
      next: items.length > 0 ? items[items.length - 1].alertSince : ''
    }
  }

  // 获取告警详情
  static async getAlertDetail(issueID: string): Promise<AlertDetail> {
    await delay(300) // 模拟网络延迟
    console.log(`Mock API: 获取告警详情 - issueID: ${issueID}`)
    
    const alertDetail = mockAlertDetails[issueID]
    if (!alertDetail) {
      throw new Error(`告警详情不存在: ${issueID}`)
    }
    
    return alertDetail
  }
}

// 导出Mock API实例
export const mockApi = MockApiService
