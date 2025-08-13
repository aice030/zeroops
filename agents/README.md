# Agents 包文档

## 概述

`agents` 包是一个基于认知内核的多步智能体系统，灵感来源于 smolagents。该包提供了完整的智能体框架，包括模型调用、工具管理、会话管理、评估等功能。

## 包结构

```
agents/
├── __init__.py          # 包初始化文件
├── agent.py            # 核心智能体类
├── model.py            # LLM 模型包装器
├── session.py          # 会话管理
├── tool.py             # 工具系统
├── utils.py            # 工具函数
├── evaluator.py        # 智能体评估器
├── evaluator_prompt.py # 评估提示词
└── gaia_scorer.py      # Gaia 基准评分器
```

## 文件详细说明

### 1. `__init__.py`

**功能**: 包初始化文件
**内容**: 
- 包含包的描述信息
- 灵感来源于 smolagents

### 2. `agent.py` - 核心智能体系统

#### 主要类

##### `AgentResult`
**功能**: 智能体调用的结果存储
**数据结构**:
```python
class AgentResult:
    output: str      # 格式化的输出
    log: str         # 其他输出信息
    task: str        # 目标任务
    repr: str        # 显式表示
```
**方法**:
- `to_dict()`: 转换为字典
- `__contains__()`: 检查键是否存在
- `__getitem__()`: 字典式访问
- `__repr__()`: 字符串表示

##### `ActionResult`
**功能**: 单个动作的结果
**数据结构**:
```python
class ActionResult:
    action: str      # 动作名称
    result: str      # 动作结果
```

##### `MultiStepAgent`
**功能**: 多步智能体的核心类
**数据结构**:
```python
class MultiStepAgent:
    name: str                    # 智能体名称
    description: str             # 智能体描述
    sub_agent_names: List[str]   # 子智能体名称列表
    tools: List[Tool]           # 工具列表
    model: LLM                  # 主循环模型
    templates: dict             # 模板名称
    max_steps: int              # 最大步数
    max_time_limit: int         # 时间限制
    recent_steps: int           # 最近步数
    store_io: bool              # 是否存储输入输出
    exec_timeout_with_call: int # 带调用的执行超时
    exec_timeout_wo_call: int   # 不带调用的执行超时
    obs_max_token: int          # 观察最大token数
    active_functions: List[str] # 活跃函数列表
```

**核心方法**:
- `__call__()`: 作为托管智能体调用
- `run()`: 运行主智能体
- `yield_session_run()`: 主运行循环
- `step()`: 执行单个步骤
- `step_check_end()`: 检查是否结束
- `finalize()`: 最终化处理

**停止原因**:
- `NORMAL_END`: 正常结束
- `MAX_STEP`: 超过最大步数
- `MAX_TIME`: 超过时间限制

**模板系统**:
- 支持模板注册和获取
- 使用 `TemplatedString` 进行模板管理
- 支持动态模板替换

### 3. `model.py` - LLM 模型包装器

#### 主要类

##### `OpenaiHelper`
**功能**: OpenAI API 助手，支持多种模型和端点
**特性**:
- 支持七牛云、OpenAI官方等多种端点
- 模型别名映射和家族默认参数
- 智能端点检测和缓存管理

**支持的模型**:
- **开源大模型**: `gpt-oss-120b`, `gpt-oss-20b`
- **DeepSeek**: `deepseek-v3`
- **GLM**: `glm-4.5`, `glm-4.5-air`
- **Kimi**: `kimi-k2`
- **Qwen**: `qwen-turbo`, `qwen-max-2025-01-25`, `qwen3-32b`
- **MiniMax**: `MiniMax-M1`
- **NVIDIA**: `nvidia/llama-3.3-nemotron-super-49b-v1.5`

**模型家族默认参数**:
- `deepseek`: temperature=0.2, top_p=0.9
- `glm`: temperature=0.3
- `kimi/qwen/llama/nemotron/minimax/gpt-oss`: temperature=0.2

**方法**:
- `get_openai_client()`: 获取OpenAI客户端
- `call_chat()`: 调用聊天接口
- `resolve_model()`: 解析模型别名
- `_merge_family_defaults()`: 合并家族默认参数

##### `LLM`
**功能**: 统一的LLM接口
**数据结构**:
```python
class LLM:
    call_target: str            # 调用目标 (gpt:model_name, manual, fake, http...)
    thinking: bool              # 是否启用推理模式
    print_call_in: str          # 输入打印样式
    print_call_out: str         # 输出打印样式
    max_retry_times: int        # 最大重试次数
    seed: int                   # 随机种子
    request_timeout: int        # 请求超时时间
    max_token_num: int          # 最大token数
    call_kwargs: dict           # 调用参数
```

**支持的调用目标类型**:
- `"manual"`: 手动输入模式
- `"fake"`: 假响应模式，返回固定文本
- `"gpt:model_name"`: OpenAI GPT模型
- `"http://..."`: HTTP API端点

**方法**:
- `__call__()`: 调用LLM
- `get_call_stat()`: 获取调用统计
- `set_seed()`: 设置随机种子
- `_call_openai_chat()`: 调用OpenAI聊天接口
- `_call_with_messages()`: 统一的消息调用接口

### 4. `session.py` - 会话管理

#### 主要类

##### `AgentSession`
**功能**: 单个任务运行的会话
**数据结构**:
```python
class AgentSession:
    id: str           # 会话ID
    info: dict        # 会话信息
    task: str         # 目标任务
    steps: List[dict] # 步骤列表
```

**方法**:
- `to_dict()`: 转换为字典
- `from_dict()`: 从字典恢复
- `num_of_steps()`: 获取步数
- `get_current_step()`: 获取当前步骤
- `get_specific_step()`: 获取特定步骤
- `get_latest_steps()`: 获取最近步骤
- `add_step()`: 添加步骤

### 5. `tool.py` - 工具系统

#### 主要类

##### `Tool`
**功能**: 工具基类
**数据结构**:
```python
class Tool:
    name: str  # 工具名称
```

**抽象方法**:
- `get_function_definition()`: 获取函数定义
- `__call__()`: 执行工具

##### `StopTool`
**功能**: 停止工具，用于任务完成
**数据结构**:
```python
class StopTool(Tool):
    agent: MultiStepAgent  # 关联的智能体
```

**方法**:
- `get_function_definition()`: 返回停止函数的定义
- `__call__()`: 执行停止操作，返回 `StopResult`

##### `AskLLMTool`
**功能**: 直接查询LLM的工具
**数据结构**:
```python
class AskLLMTool(Tool):
    llm: LLM  # LLM实例
```

**方法**:
- `set_llm()`: 设置LLM
- `get_function_definition()`: 返回查询函数的定义
- `__call__()`: 执行LLM查询

##### `SimpleSearchTool`
**功能**: 简单搜索工具
**数据结构**:
```python
class SimpleSearchTool(Tool):
    target: str      # 搜索目标
    llm: LLM        # LLM实例
    max_results: int # 最大结果数
    list_enum: bool  # 是否枚举列表
```

**特性**:
- 支持多种搜索后端（DuckDuckGo等）
- 可配置最大结果数量
- 支持列表枚举模式

### 6. `utils.py` - 工具函数

#### 主要功能

##### 打印和日志
- `rprint()`: 富文本打印，支持样式和颜色
- `zlog()`: 日志打印别名
- `zwarn()`: 警告打印，红色背景

##### JSON处理
- `MyJsonEncoder`: 自定义JSON编码器，支持to_dict方法
- `my_json_dumps()`: JSON序列化
- `tuple_keys_to_str()`: 元组键转字符串

##### 重试机制
- `wrapped_trying()`: 包装重试函数，支持指数退避

##### 环境变量
- `GET_ENV_VAR()`: 获取环境变量，支持多个备选键
- `load_dotenv()`: 自动加载.env文件

##### 基础类
- `KwargsInitializable`: 支持kwargs初始化的基类
- `TemplatedString`: 模板字符串处理，支持f-string语法
- `CodeExecutor`: 代码执行器
- `WithWrapper`: 上下文管理器包装器

##### 其他工具
- `get_unique_id()`: 生成唯一ID
- `my_open_with()`: 文件操作包装器

### 7. `evaluator.py` - 智能体评估器

#### 主要功能

##### 评估函数
- `get_prompt()`: 加载系统提示词
- `rule_filter_final_action_message()`: 过滤最终动作消息
- `rule_filter_end_message()`: 过滤结束消息
- `remove_keys()`: 递归移除指定键
- `get_messages()`: 构造消息列表

##### 评估逻辑
- 支持多种评估规则
- 自定义链式思维问答评估器
- 支持多模态内容评估
- 集成多个答案的集成评估

##### 主要方法
- `summarize()`: 总结智能体执行轨迹
- `custom_cot_qa_evaluate()`: 自定义链式思维评估
- `gpt_judge()`: GPT判断答案正确性
- `detect_failure()`: 检测失败情况
- `ensemble()`: 集成多个答案
- `worker_detect_ask_llm()`: 检测是否需要询问LLM

##### 环境变量配置
- `EVALUATOR_LLM`: 评估器使用的模型，默认为 `"fake"`
- 支持 `"manual"`, `"fake"`, `"gpt:model_name"`, `"http://..."` 等模式

### 8. `evaluator_prompt.py` - 评估提示词

#### 内容
包含预定义的评估提示词：

**`gpt_judge_heuristic_with_traj`**: GPT判断启发式轨迹
- 验证自动化智能体响应的正确性
- 基于给定规则检查非负性、合理性、成功性和可靠性
- 返回 `==yes==` 或 `==no==`

**`gpt_chooser`**: GPT选择器
- 评估多个智能体解决方案
- 选择正确的解决方案
- 返回解决方案索引或-1

**`ask_llm_system_prompt`**: 询问LLM系统提示词
- 判断是否需要调用ask_llm函数
- 基于之前的失败或信息不足情况
- 返回1（需要调用）或0（不需要调用）

### 9. `gaia_scorer.py` - Gaia 基准评分器

#### 主要功能
基于 [Gaia Benchmark](https://huggingface.co/spaces/gaia-benchmark/leaderboard) 的评分系统

**核心函数**:
- `question_scorer()`: 主要评分函数
- `normalize_number_str()`: 数字字符串标准化
- `split_string()`: 字符串分割
- `normalize_str()`: 字符串标准化

**评分逻辑**:
- **数字评分**: 支持货币、百分比等格式，移除标点符号后比较
- **列表评分**: 支持逗号、分号分隔的列表，逐元素比较
- **字符串评分**: 标准化后精确匹配

**特性**:
- 自动检测数据类型（数字、列表、字符串）
- 支持多种分隔符
- 智能字符串标准化
- 详细的评分日志

## 包依赖关系

### 依赖层次结构

```
agents/
├── __init__.py (无依赖)
├── utils.py (无外部依赖，基础工具)
├── session.py (依赖: utils)
├── tool.py (依赖: utils)
├── model.py (依赖: utils)
├── agent.py (依赖: model, session, tool, utils)
├── evaluator.py (依赖: utils, model, evaluator_prompt, gaia_scorer)
├── evaluator_prompt.py (无依赖)
└── gaia_scorer.py (无依赖)
```

### 详细依赖关系

1. **基础层**
   - `utils.py`: 提供基础工具类和函数，被其他所有模块依赖

2. **核心层**
   - `session.py`: 依赖 `utils.py` 中的 `get_unique_id`
   - `tool.py`: 依赖 `utils.py` 中的 `KwargsInitializable`, `rprint`, `GET_ENV_VAR`
   - `model.py`: 依赖 `utils.py` 中的 `wrapped_trying`, `rprint`, `GET_ENV_VAR`, `KwargsInitializable`

3. **智能体层**
   - `agent.py`: 依赖 `model.py`, `session.py`, `tool.py`, `utils.py`

4. **评估层**
   - `evaluator_prompt.py`: 独立模块，提供提示词
   - `gaia_scorer.py`: 独立模块，提供评分功能
   - `evaluator.py`: 依赖多个模块，包括 `gaia_scorer`

### 外部依赖

- `transformers`: 用于token计算和模型处理
- `requests`: HTTP请求
- `openai`: OpenAI API客户端
- `rich`: 富文本控制台输出
- `numpy`: 数值计算

## 使用流程

1. **初始化**: 创建 `MultiStepAgent` 实例
2. **配置**: 设置模型、工具、模板等
3. **运行**: 调用 `run()` 方法执行任务
4. **监控**: 通过会话对象监控执行过程
5. **评估**: 使用评估器评估结果质量

## 配置说明

### 环境变量配置

```bash
# 评估器模型配置
export EVALUATOR_LLM=gpt:gpt-4o-mini
export EVALUATOR_LLM=gpt:deepseek-v3
export EVALUATOR_LLM=gpt:qwen-max-2025-01-25
export EVALUATOR_LLM=fake  # 使用假响应模式（默认）
export EVALUATOR_LLM=manual  # 使用手动输入模式

# OpenAI API配置
export OPENAI_API_KEY=your-api-key
export OPENAI_ENDPOINT=https://your-endpoint.com/v1

# 七牛云配置
export QINIU_API_KEY=your-qiniu-key
export QINIU_ENDPOINT=https://your-qiniu-endpoint.com/v1

# 搜索后端配置
export SEARCH_BACKEND=DuckDuckGo

# 其他配置
export NO_FORCE_TERMINAL=false
```

### 模型调用格式

```python
# 使用预定义模型
llm = LLM(call_target="gpt:gpt-4o-mini")
llm = LLM(call_target="gpt:deepseek-v3")
llm = LLM(call_target="gpt:qwen-max-2025-01-25")

# 使用特殊模式
llm = LLM(call_target="fake")      # 假响应模式
llm = LLM(call_target="manual")    # 手动输入模式

# 使用自定义端点
llm = LLM(call_target="http://your-api.com/v1")
```

### 工具配置

```python
# 创建停止工具
stop_tool = StopTool(agent=agent)

# 创建LLM查询工具
ask_tool = AskLLMTool(llm=llm)

# 创建搜索工具
search_tool = SimpleSearchTool(
    target="DuckDuckGo",
    llm=llm,
    max_results=10,
    list_enum=True
)
```

## 设计模式

- **模板模式**: 使用模板字符串进行提示词管理
- **策略模式**: 不同的工具和模型实现
- **观察者模式**: 会话状态监控
- **工厂模式**: 工具和模型的创建
- **装饰器模式**: 重试和超时包装
- **组合模式**: 工具和子智能体的组合

## 扩展性

### 添加新模型
在 `OpenaiHelper.MODEL_ALIASES` 中添加新模型别名：
```python
MODEL_ALIASES = {
    "your-model": "your-actual-model-id",
    # ... 其他模型
}
```

### 添加新工具
继承 `Tool` 基类并实现必要方法：
```python
class YourTool(Tool):
    def __init__(self, **kwargs):
        super().__init__(name="your_tool")
        # 初始化代码
    
    def get_function_definition(self, short: bool):
        # 返回工具定义
    
    def __call__(self, *args, **kwargs):
        # 执行工具逻辑
```

### 添加新评估方法
在 `Evaluator` 类中添加新的评估方法：
```python
def your_evaluation_method(self, data):
    # 实现评估逻辑
    pass
```

### 添加新评分器
在 `gaia_scorer.py` 中添加新的评分函数：
```python
def your_scorer(model_answer: str, ground_truth: str) -> bool:
    # 实现评分逻辑
    pass
```

## 故障排除

### 常见问题

1. **导入错误**: 检查相对导入路径是否正确
2. **环境变量未设置**: 确保设置了必要的环境变量
3. **模型调用失败**: 检查API密钥和网络连接
4. **工具执行错误**: 验证工具配置和依赖

### 调试技巧

- 使用 `fake` 模式进行快速测试
- 启用详细日志输出
- 检查会话状态和步骤记录
- 使用 `manual` 模式进行交互式调试

## 总结

这个包提供了一个完整的智能体框架，支持多步推理、工具使用、会话管理和结果评估。主要特点包括：

- **模块化设计**: 清晰的依赖层次和模块分离
- **灵活配置**: 支持多种模型、工具和评估方式
- **易于扩展**: 基于抽象基类的可扩展架构
- **完整功能**: 从智能体执行到结果评估的完整流程
- **生产就绪**: 包含错误处理、重试机制和超时控制

是一个功能丰富的认知内核智能体系统，适合构建复杂的AI应用和自动化任务。
