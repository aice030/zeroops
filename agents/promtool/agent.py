#

import time
import re
import random
from functools import partial
import multiprocessing as mp

from ..base.agent import MultiStepAgent, register_template, AgentResult
from ..base.tool import StopTool, AskLLMTool, SimpleSearchTool, QueryErrorLogsTool, QueryDependencyTool
from ..base.utils import zwarn, GET_ENV_VAR
from ..promtool.agent import Prom_ToolAgent
# Removed web_agent imports
# Removed file_agent imports
from .prompts import PROMPTS as ZO_PROMPTS

# --
class ZOAgent(MultiStepAgent):
    def __init__(self, **kwargs):
        # note: this is a little tricky since things will get re-init again in super().__init__
        # sub-agents - removed web_agent and file_agent initialization
        self.tool_ask_llm = AskLLMTool()
        self.tool_simple_search = SimpleSearchTool()
        self.tool_query_error_logs = QueryErrorLogsTool()
        self.tool_query_dependency = QueryDependencyTool()
        self.prom_agent = Prom_ToolAgent()
        feed_kwargs = dict(
            name="zo_agent",
            description="Cognitive Kernel, an initial autopilot system.",
            templates={"plan": "zo_plan", "action": "zo_action", "end": "zo_end", "aggr": "zo_aggr"},  # template names (no need of END here since we do NOT use __call__ for this)
            active_functions=["stop", "ask_llm", "simple_web_search", "query_error_logs", "query_dependency", "prom_agent"],  # removed web_agent and file_agent
            sub_agent_names=["prom_agent"],  # removed web_agent and file_agent
            tools=[StopTool(agent=self), self.tool_ask_llm, self.tool_simple_search, self.tool_query_error_logs, self.tool_query_dependency],  # add related tools
            max_steps=16,  # still give it more steps
            max_time_limit=4200,  # 70 minutes
            exec_timeout_with_call=1000,  # if calling sub-agent
            exec_timeout_wo_call=200,  # if not calling sub-agent
        )
        feed_kwargs.update(kwargs)
        # our new args
        self.step_mrun = 1  # step-level multiple run to do ensemble
        self.mrun_pool_size = 5  # max pool size for parallel running
        self.mrun_multimodal_count = 0  # how many runs to go with multimodal-web
        # --
        register_template(ZO_PROMPTS)  # add web prompts
        super().__init__(**feed_kwargs)
        self.tool_ask_llm.set_llm(self.model)  # another tricky part, we need to assign LLM later
        self.tool_simple_search.set_llm(self.model)
        self.tool_query_error_logs.set_llm(self.model)
        self.tool_query_dependency.set_llm(self.model)
        # --

    def get_function_definition(self, short: bool):
        raise RuntimeError("Should NOT use ZOAgent as a sub-agent!")

    def _super_step_action(self, _id: int, need_sleep: bool, action_res, action_input_kwargs, **kwargs):
        if need_sleep and _id:
            time.sleep(5 * int(_id))  # do not run them all at once!
        # --
        if _id is None:  # not multiple run mode
            ret = super().step_action(action_res, action_input_kwargs, **kwargs)
        else:
            # Removed web_agent multimodal handling
            _old_seed = self.get_seed()
            _new_seed = _old_seed + int(_id)
            try:
                self.set_seed(_new_seed)
                ret = super().step_action(action_res, action_input_kwargs, **kwargs)
            finally:
                self.set_seed(_old_seed)
        # --
        return ret

    def step_action(self, action_res, action_input_kwargs, **kwargs):
        _need_multiple = any(f"{kk}(" in action_res["code"] for kk in ["ask_llm", "prom_agent"])  # removed web_agent and file_agent
        if self.step_mrun <= 1 or (not _need_multiple):  # just run once
            return self._super_step_action(None, False, action_res, action_input_kwargs, **kwargs)
        else:  # multiple run and aggregation
            _need_sleep = False  # removed web_agent sleep logic
            with mp.Pool(min(self.mrun_pool_size, self.step_mrun)) as pool:  # note: no handle of errors here since the wraps (including the timeout) will be inside each sub-process
                # all_results = pool.map(partial(self._super_step_action, need_sleep=_need_sleep, action_res=action_res, action_input_kwargs=action_input_kwargs, **kwargs), list(range(self.step_mrun)))
                all_results = pool.map(zo_step_action, [(self, _id, _need_sleep, action_res, action_input_kwargs, kwargs) for _id in range(self.step_mrun)])
            # aggregate results
            aggr_res = None
            try:
                _aggr_inputs = action_input_kwargs.copy()
                _aggr_inputs["current_step"] = f"Thought: {action_res.get('thought')}\nAction: ```\n{action_res.get('code')}```"
                _aggr_inputs["result_list"] = "\n".join([f"### Result {ii}\n{rr}\n" for ii, rr in enumerate(all_results)])
                aggr_messages = self.templates["aggr"].format(**_aggr_inputs)
                aggr_response = self.step_call(messages=aggr_messages, session=None)  # note: for simplicity no need session info here for aggr!
                aggr_res = self._parse_output(aggr_response)
                if self.store_io:  # further storage
                    aggr_res.update({"llm_input": aggr_messages, "llm_output": aggr_response})
                _idx_str = re.findall(r"print\(.*?(\d+).*?\)", aggr_res["code"])
                _sel = int(_idx_str[-1])
                assert _sel >= 0 and _sel < len(all_results), f"Out of index error for selection index {_sel}"  # detect out of index error!
            except Exception as e:
                zwarn(f"Error when doing selection: {aggr_res} -> {e}")
                _sel = 0  # simply select the first one
            _ret = AgentResult(repr=repr(all_results[_sel]), sel_aggr=aggr_res, sel_cands=all_results, sel_idx=_sel)  # store all the information!
            return _ret
        # --

# --
# make it a top-level function
def zo_step_action(args):
    zo, _id, need_sleep, action_res, action_input_kwargs, kwargs = args
    return zo._super_step_action(_id, need_sleep, action_res, action_input_kwargs, **kwargs)
# --