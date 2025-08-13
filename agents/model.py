#

# a simple wrapper for LLM calling

import time
import requests
from .utils import wrapped_trying, rprint, GET_ENV_VAR, KwargsInitializable

# helper
def update_stat(stat, call_return):
    if stat is None:
        return
    usage = call_return.get("usage", {}) or {}
    key_map = {
        "prompt_tokens": "prompt_tokens",
        "completion_tokens": "completion_tokens",
        "total_tokens": "total_tokens",
        "input_tokens": "prompt_tokens",
        "output_tokens": "completion_tokens",
        "inputTokens": "prompt_tokens",
        "outputTokens": "completion_tokens",
        "totalTokens": "total_tokens",
    }
    if usage:
        stat["llm_call"] = stat.get("llm_call", 0) + 1
        for k, v in usage.items():
            if isinstance(v, (int, float)):
                norm = key_map.get(k)
                if norm:
                    stat[norm] = stat.get(norm, 0) + int(v)



import re

class OpenaiHelper:
    _openai_clients = {}  # key: (vendor, model_name, endpoint) -> client

    # —— 规范化模型名作为 env 后缀：把非 [A-Za-z0-9_] 全换成 _
    @staticmethod
    def _env_suffix(model_name: str) -> str:
        if not model_name:
            return ""
        return "_" + re.sub(r"[^A-Za-z0-9_]", "_", model_name)

    # —— QINIU 推荐主/备接入点，没配 QINIU_ENDPOINT 时使用
    QINIU_DEFAULT_ENDPOINT = "https://openai.qiniu.com/v1"
    QINIU_FALLBACK_ENDPOINT = "https://api.qnaigc.com/v1"

    # —— 你的名单：友好名 -> 实际模型ID（目前多数等于自身；slash 型也支持）
    MODEL_ALIASES = {
        # 你列出来的
        "gpt-oss-120b": "gpt-oss-120b",
        "gpt-oss-20b": "gpt-oss-20b",
        "deepseek-v3": "deepseek-v3",
        "glm-4.5": "glm-4.5",
        "glm-4.5-air": "glm-4.5-air",
        "kimi-k2": "kimi-k2",
        "qwen-turbo": "qwen-turbo",
        "MiniMax-M1": "MiniMax-M1",
        "nvidia/llama-3.3-nemotron-super-49b-v1.5": "nvidia/llama-3.3-nemotron-super-49b-v1.5",
        "qwen-max-2025-01-25": "qwen-max-2025-01-25",
        "qwen3-32b": "qwen3-32b",
        # 如需自定义本地别名，可继续添加： "nv-nemotron-49b": "nvidia/llama-3.3-nemotron-super-49b-v1.5",
    }

    @staticmethod
    def resolve_model(model: str) -> str:
        """把友好名映射到真实模型ID；默认原样返回。"""
        return OpenaiHelper.MODEL_ALIASES.get(model, model)

    # —— 根据“模型家族”设默认参数（按前缀/关键词匹配）
    FAMILY_DEFAULTS = [
        (r"^deepseek",           {"temperature": 0.2, "top_p": 0.9}),
        (r"^glm",                {"temperature": 0.3}),
        (r"^kimi",               {"temperature": 0.2}),
        (r"^qwen",               {"temperature": 0.2}),
        (r"(^llama|nemotron)",   {"temperature": 0.2}),
        (r"^minimax",            {"temperature": 0.2}),
        (r"^gpt-oss",            {"temperature": 0.2}),
        # 需要 reasoning 的模型可以在这里加自定义字段（若兼容层支持）
        # (r"^deepseek-.*r1",    {"reasoning": {"effort": "medium"}}),
    ]

    @staticmethod
    def _merge_family_defaults(model: str, kwargs: dict) -> dict:
        merged = dict(kwargs or {})
        lower = (model or "").lower()
        for pattern, defaults in OpenaiHelper.FAMILY_DEFAULTS:
            if re.search(pattern, lower):
                for k, v in defaults.items():
                    merged.setdefault(k, v)
                break
        return merged

    @staticmethod
    def _ensure_v1(url: str) -> str:
        if not url:
            return url
        u = url.rstrip("/")
        return u if u.endswith("/v1") else (u + "/v1")

    @staticmethod
    def _resolve_vendor(model_name_suffix=""):
        # 1) QINIU（检测 QINIU_API_KEY/_ENDPOINT） 2) OpenAI / 其他兼容（OPENAI_*）
        if (GET_ENV_VAR("QINIU_API_KEY", f"QINIU_API_KEY{model_name_suffix}") or
            GET_ENV_VAR("QINIU_ENDPOINT", f"QINIU_ENDPOINT{model_name_suffix}")):
            return "qiniu"
        return "openai"

    @staticmethod
    def get_openai_client(model_name="", api_endpoint="", api_key=""):
        import openai
        # ★ 用“规范化后缀”避免 -, / 之类导致 env 读取失败
        model_suffix = OpenaiHelper._env_suffix(model_name)
        vendor = OpenaiHelper._resolve_vendor(model_suffix)

        if vendor == "qiniu":
            endpoint = GET_ENV_VAR("QINIU_ENDPOINT", f"QINIU_ENDPOINT{model_suffix}", df=api_endpoint)
            key = GET_ENV_VAR("QINIU_API_KEY", f"QINIU_API_KEY{model_suffix}", df=api_key)

            if not endpoint:
                # 没配就用默认主域名；如遇网络问题可切备用域名
                endpoint = OpenaiHelper.QINIU_DEFAULT_ENDPOINT
            endpoint = OpenaiHelper._ensure_v1(endpoint)

            cache_key = (vendor, model_name, endpoint)
            if cache_key not in OpenaiHelper._openai_clients:
                OpenaiHelper._openai_clients[cache_key] = openai.OpenAI(base_url=endpoint, api_key=key)
            return OpenaiHelper._openai_clients[cache_key]

        # 默认：OpenAI 官方或其它兼容（OPENAI_ENDPOINT 存在即兼容）
        endpoint = GET_ENV_VAR("OPENAI_ENDPOINT", f"OPENAI_ENDPOINT{model_suffix}", df=api_endpoint)
        key = GET_ENV_VAR("OPENAI_API_KEY", f"OPENAI_API_KEY{model_suffix}", df=api_key)
        if endpoint:
            endpoint = OpenaiHelper._ensure_v1(endpoint)
        cache_key = ("openai", model_name, endpoint or "official")
        if cache_key not in OpenaiHelper._openai_clients:
            if endpoint:
                OpenaiHelper._openai_clients[cache_key] = openai.OpenAI(base_url=endpoint, api_key=key)
            else:
                OpenaiHelper._openai_clients[cache_key] = openai.OpenAI(api_key=key)
        return OpenaiHelper._openai_clients[cache_key]


    @staticmethod
    def call_chat(messages, stat=None, **openai_kwargs):
        # 1) 解析模型别名
        raw_model = openai_kwargs.get("model", "")
        model = OpenaiHelper.resolve_model(raw_model)
        openai_kwargs["model"] = model

        # 2) 合并家族默认参数（未显式传入的才填充）
        openai_kwargs = OpenaiHelper._merge_family_defaults(model, openai_kwargs)

        rprint(f"Call gpt with openai_kwargs={openai_kwargs}")

        _client = OpenaiHelper.get_openai_client(model)
        chat_completion = _client.chat.completions.create(messages=messages, **openai_kwargs)

        call_return = chat_completion.to_dict() if hasattr(chat_completion, "to_dict") else chat_completion

        update_stat(stat, call_return)

        # 提取文本
        try:
            response = call_return["choices"][0]["message"].get("content") or ""
        except Exception:
            response = ""

        if response.strip() == "":
            raise RuntimeError(f"Get empty response from model: {call_return}")
        return response

class LLM(KwargsInitializable):
    def __init__(self, **kwargs):
        # basics
        self.call_target = "manual"  # fake=fake, manual=input, gpt(gpt:model_name)=openai [such as gpt:gpt-4o-mini], request(http...)=request
        self.thinking = False
        self.print_call_in = "white on blue"  # easier to read
        self.print_call_out = "white on green"  # easier to read
        self.max_retry_times = 5  # <0 means always trying
        self.seed = 1377  # zero means no seed!
        # request
        self.request_timeout = 100  # timeout time
        self.max_token_num = 20000
        self.call_kwargs = {"temperature": 0.0, "top_p": 0.95, "max_tokens": 4096}  # other kwargs for gpt/request calling
        # --
        super().__init__(**kwargs)  # init
        # --
        # post init
        self.call_target_type = self.get_call_target_type()
        self.call_stat = {}  # stat of calling

    def __repr__(self):
        return f"LLM(target={self.call_target},kwargs={self.call_kwargs})"

    def get_seed(self):
        return self.seed

    def set_seed(self, seed):
        self.seed = seed

    def __call__(self, messages, **kwargs):
        func = lambda: self._call_with_messages(messages, **kwargs)
        return wrapped_trying(func, max_times=self.max_retry_times)

    def get_call_stat(self, clear=False):
        ret = self.call_stat.copy()
        if clear:  # clear stat
            self.clear_call_stat()
        return ret

    def clear_call_stat(self):
        self.call_stat.clear()

    def get_call_target_type(self):
        _trg = self.call_target
        if _trg == "manual":
            return "manual"
        elif _trg == "fake":
            return "fake"
        elif _trg.startswith("gpt:"):
            return "gpt"
        elif _trg.startswith("http"):
            return "request"
        else:
            raise RuntimeError(f"UNK call_target = {_trg}")

    def show_messages_str(self, messages, calling_kwargs, rprint_style):
        ret_ss = []
        if isinstance(messages, list):
            for one_mesg in messages:
                _content = one_mesg['content']
                if isinstance(_content, list):
                    _content = "\n\n".join([(z['text'] if z['type']=='text' else f"<{str(z)[:150]}...>") for z in _content])
                ret_ss.extend([f"=====\n", (f"{one_mesg['role']}: {_content}\n", rprint_style)])
        else:
            ret_ss.append((f"{messages}\n", rprint_style))
        ret = [f"### ----- Call {self.call_target} with {calling_kwargs} [ctime={time.ctime()}]\n{'#'*10}\n"] + ret_ss + [f"{'#'*10}"]
        return ret

    # still return a str here, for simplicity!
    def _call_with_messages(self, messages, **kwargs):
        time0 = time.perf_counter()
        _call_target_type = self.call_target_type
        _call_kwargs = self.call_kwargs.copy()
        _call_kwargs.update(kwargs)  # this time's kwargs
        if self.print_call_in:
            rprint(self.show_messages_str(messages, _call_kwargs, self.print_call_in))  # print it out
        # --
        if _call_target_type == "manual":
            user_input = input("Put your input >> ")
            response = user_input.strip()
            ret = response
        elif _call_target_type == "fake":
            ret = "You are correct! As long as you are happy!"
        elif _call_target_type == "gpt":
            ret = self._call_openai_chat(messages, **_call_kwargs)
        elif _call_target_type == "request":
            headers = {"Content-Type": "application/json"}
            if isinstance(messages, list):
                json_data = {
                    "model": "ck",
                    "stop": ["<|eot_id|>", "<|eom_id|>", "<|im_end|>"],
                    "messages": messages,
                }
                if self.seed != 0:  # only if non-zero!
                    json_data.update(seed=self.seed)
            else:  # directly put it!
                json_data = messages.copy()
            json_data.update(_call_kwargs)
            r = requests.post(self.call_target, headers=headers, json=json_data, timeout=self.request_timeout)
            assert (200 <= r.status_code <= 300), f"response error: {r.status_code} {json_data}"
            call_return = r.json()
            if isinstance(call_return, dict) and "choices" in call_return:
                update_stat(self.call_stat, call_return)
                ret0 = call_return["choices"][0]
                if "message" in ret0:
                    ret = ret0["message"]["content"]  # chat-format
                    # thought = ret0["message"]["reasoning_content"] # for qwen3
                    # remove <think> </think> tokens
                    import re
                    ret = re.sub(r'<think>.*?</think>', '', ret, flags=re.DOTALL)
                else:
                    ret = ret0["text"]
            else:  # directly return the full object
                ret = call_return
        else:
            ret = None
        # --
        assert ret is not None, f"Calling failed for {_call_target_type}"
        if self.print_call_out:
            ss = [f"# == Calling result [ctime={time.ctime()}, interval={time.perf_counter() - time0:.3f}s] =>\n", (ret, self.print_call_out), "\n# =="]
            rprint(ss)
        return ret

    def _call_openai_chat(self, messages, **kwargs):
        _gpt_kwargs = {"model": self.call_target.split(":", 1)[1]}
        _gpt_kwargs.update(kwargs)

        # Message truncation removed - messages will be sent as-is

        while True:
            try:
                ret = OpenaiHelper.call_chat(messages, stat=self.call_stat, **_gpt_kwargs)
                return ret
            except Exception as e:  # simply catch everything!
                rprint(f"Get error when calling gpt: {e}", style="white on red")
                if type(e).__name__ in ["RateLimitError"]:
                    time.sleep(10)
                elif type(e).__name__ == "BadRequestError":
                    error_str = str(e)
                    if "ResponsibleAIPolicyViolation" in error_str or "content_filter" in error_str:
                        # rprint("Jailbreak or content filter violation detected. Please modify your prompt.", style="white on red")
                        return "Thought: Jailbreak or content filter violation detected. Please modify your prompt or stop with N/A."
                    else:
                        rprint(f"BadRequestError: {error_str}", style="white on red")
                    break
                else:
                    break
        return None


# --
def test_llm():
    # Check if API key is available
    api_key = GET_ENV_VAR("OPENAI_API_KEY", "QINIU_API_KEY")
    if not api_key or api_key == "your_openai_api_key_here":
        print("⚠️  Warning: No valid API key found!")
        print("Please set your API key in the .env file or environment variables.")
        print("You can copy env_example.txt to .env and fill in your actual API key.")
        print("\nFor testing purposes, switching to 'fake' mode...")
        llm = LLM(call_target="fake")
    else:
        print(f"✅ API key found: {api_key[:8]}...")
        llm = LLM(call_target="gpt:gpt-oss-20b")
    
    messages = [{"role": "system", "content": "You are a helpful assistant."}]
    while True:
        try:
            p = input("Prompt >> ")
            if p.lower() in ['quit', 'exit', 'q']:
                break
            messages.append({"role": "user", "content": p.strip()})
            r = llm(messages)
            messages.append({"role": "assistant", "content": r})
        except KeyboardInterrupt:
            print("\nExiting...")
            break
        except Exception as e:
            print(f"Error: {e}")
            break

if __name__ == '__main__':
    test_llm()
