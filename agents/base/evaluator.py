#

# a simple wrapper for LLM calling

import time
import requests
from copy import deepcopy
import time
from .utils import wrapped_trying, rprint, GET_ENV_VAR, KwargsInitializable
import re
from openai import OpenAIError
from .gaia_scorer import question_scorer
from .model import OpenaiHelper, LLM
from .evaluator_prompt import prompt_dict
import os
import json
from collections import Counter


def get_prompt(prompt_name):
    """Loads the system prompt from the given JSON file."""
    return prompt_dict[prompt_name]
    


# filter out final action message and end message if they are None.
def rule_filter_final_action_message(final_message):
    if not 'stop' in final_message:
        return True
    patterns = [r'stop.*not found', r'stop.*none', r'stop.*\'\'', r'stop.*""']
    return any(bool(re.search(p, final_message, re.IGNORECASE | re.DOTALL)) for p in patterns)

def rule_filter_end_message(end_message):
    patterns = [r'output.*not found', r'output.*none', r'output.*\'\'', r'output.*""']
    return any(bool(re.search(p, end_message, re.IGNORECASE | re.DOTALL)) for p in patterns)


def remove_keys(d, keys_to_remove=["boxed_screenshot", "llm_input", "state", "llm_output", "plan", "info", "snapshot", "browser_id", "page_id", "_orig", "current_has_cookie_popup", "expanded_part", "curr_step", "curr_step", "curr_screenshot_mode", "total_actual_step", "num_revert_state", "answer", "file_state_before", "web_state_before"]):
    """
    Recursively removes specified keys from a nested dictionary.
    
    Parameters:
    - d (dict): The input dictionary.
    - keys_to_remove (list): List of keys to remove from the dictionary at all levels.
    
    Returns:
    - dict: A new dictionary with the specified keys removed.
    """
    if not isinstance(d, dict):
        return d  # If not a dictionary, return the object itself
    
    # Create a new dictionary without the keys to be removed
    result = {}
    for key, value in d.items():
        if key not in keys_to_remove:
            if isinstance(value, dict):
                result[key] = remove_keys(value, keys_to_remove)  # Recursively handle nested dictionaries
            elif isinstance(value, list):
                # If it's a list, process each item in the list
                result[key] = [remove_keys(item, keys_to_remove) if isinstance(item, dict) else item for item in value]
            else:
                result[key] = value
    return result


def get_messages(prompt, system="You are a helpful assistant.", image_urls=None):
    """
    Constructs a message list for the OpenAI API based on the provided prompt and system message.
    
    Parameters:
    - prompt (str): The user input or question.
    - system (str): The system message to guide the assistant's behavior.
    
    Returns:
    - list: A list of messages formatted for the OpenAI API.
    """
    if "gpt" in GET_ENV_VAR("EVALUATOR_LLM"):
        model = "gpt"
    else:
        model = "claude"
    if not image_urls:
        # if model == "gpt":
            return [
                {
                    "role": "system",
                    "content": system
                },
                {
                    "role": "user",
                    "content": prompt
                }
            ]
        # else:
        #     return [
        #         {
        #             "role": "developer",
        #             "content": [
        #                 {
        #                     'text': system,
        #                 },
        #             ]
        #         },
        #         {
        #             "role": "user",
        #             "content": [
        #                 {
        #                     'text': prompt,
        #                 },
        #             ]
        #         }
        #     ]
    else:
        return [
            {
                "role": "developer",
                "content": system
            },
            {
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        "text": prompt
                    },
                ] + [
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": image_url
                        }
                    } for image_url in image_urls
                ]
            }
        ]
    




class Evaluator(KwargsInitializable):
    '''
        need to configure:
            EVALUATOR_LLM - 设置评估器使用的模型，如 gpt-4o-mini, claude-3-sonnet 等
    # export EVALUATOR_LLM=gpt-4o-mini
    '''
    def __init__(self, **kwargs):
        # basics
        self.eval_method = ""
        # --
        super().__init__(**kwargs)  # init
        # 设置默认的评估器模型，如果环境变量未设置则使用 fake 模式
        evaluator_llm = GET_ENV_VAR("EVALUATOR_LLM", df="fake")
        self.helper = LLM(call_target=evaluator_llm)

        # 移除 Azure 相关代码，使用自定义评估器
        self.llm = None
        self.cot_qa_evaluator = None

    def summarize(self, inst):
        log_inst = remove_keys(inst, keys_to_remove=["boxed_screenshot", "llm_input", "state", "llm_output", "info", "snapshot", "browser_id", "page_id", "_orig", "current_has_cookie_popup", "expanded_part", "curr_step", "curr_screenshot_mode", "total_actual_step", "num_revert_state", "answer", "file_state_before", "task", "web_state_before", "eval"])
        # ref = inst["_orig"]["Annotator Metadata"]["Steps"]

        # summarize
        msg = "This is the trajectory of an agent completing a task, for each step the agent takes, please summarize the agent's action, key observations and information obtained after the action. If the trajectory includes detailed reasoning process, please also inlude the reasoning process in your summary. Your response should be in this format without any additional contents:\n\nStep 1: \nAction 1: <action>\nObservation 1: <observation>\nAction 2: <action>\nObservation 2: <observation>\n...\n\nStep 2: \nAction 1: <action>\nObservation 1: <observation>\nAction 2: <action>\nObservation 2: <observation>\n...\n\nStep 3: \nAction 1: <action>\nObservation 1: <observation>\nAction 2: <action>\nObservation 2: <observation>\n...\n\nHere is the trajectory:\n\n"
        for i in range(2):
            try:
                # log = helper.call_chat(get_messages(msg + str(log_inst)), model="gpt-4.1", max_tokens=20000, temperature=0.2, top_p=0.95, stop=None, stream=False, response_format=None)
                log = self.helper(get_messages(msg + str(log_inst)))
                break
            except Exception as e:
                print(f"Error summarizing the instance: {e}")
                log = str(log_inst)
                time.sleep(5)  # wait for a while before retrying

        return log

    def worker_detect_ask_llm(self, t):
        for step in t['steps']:
            if bool(re.search(r'ask_llm\((.*?)\)', step['action']['code'])):
                if '1' in self.determine_force_ask_llm(step['action']['thought']):
                    return True
        return False

    def determine_force_ask_llm(self, thought):
        while True:
            try:
                msg = get_messages("Thought: " + thought, system=get_prompt("ask_llm_system_prompt"))
                response = self.helper(msg)
                return response
            except Exception as e:
                print(e)
                try:
                    if any(e.body['innererror']['content_filter_result'][r]['filtered'] for r in e.body['innererror']['content_filter_result']):
                        return ''
                except:
                    pass
                if type(e).__name__ == 'RateLimitError':
                    time.sleep(10)
                elif type(e).__name__ == 'APIError':
                    time.sleep(15)
                elif type(e).__name__ == 'InvalidRequestError':
                    return ''
                else:
                    time.sleep(10)

    def custom_cot_qa_evaluate(self, item):
        """
        自定义的链式思维问答评估器，替代 LangChain 的 load_evaluator
        使用 LLM 直接评估答案质量
        """
        pred, ref, task = item['pred'], item['gold'], item['task']
        if pred is None:
            return 0
        
        # 构建评估提示
        evaluation_prompt = f"""You are an expert evaluator. Please evaluate the quality of the predicted answer compared to the reference answer.

Task: {task}
Reference Answer: {ref}
Predicted Answer: {pred}

Please rate the predicted answer on a scale of 0-10, where:
- 0: Completely incorrect or irrelevant
- 5: Partially correct but missing key information
- 10: Completely correct and comprehensive

Consider:
1. Accuracy of the information
2. Completeness of the answer
3. Relevance to the task
4. Clarity and coherence

Please respond with only a number between 0-10, no other text."""

        max_retries = 3
        for attempt in range(max_retries):
            try:
                messages = get_messages(evaluation_prompt, system="You are a fair and accurate evaluator.")
                response = self.helper(messages)
                
                # 提取评分
                try:
                    # 尝试从响应中提取数字
                    import re
                    score_match = re.search(r'(\d+(?:\.\d+)?)', response.strip())
                    if score_match:
                        score = float(score_match.group(1))
                        # 确保分数在 0-10 范围内
                        score = max(0, min(10, score))
                        return score / 10.0  # 转换为 0-1 范围
                    else:
                        # 如果没有找到数字，根据响应内容估算
                        if any(word in response.lower() for word in ['correct', 'accurate', 'perfect', 'excellent']):
                            return 0.9
                        elif any(word in response.lower() for word in ['partially', 'somewhat', 'mostly']):
                            return 0.6
                        elif any(word in response.lower() for word in ['incorrect', 'wrong', 'inaccurate']):
                            return 0.2
                        else:
                            return 0.5
                except:
                    return 0.5
                    
            except Exception as e:
                print(f"Error during custom evaluation: {e}")
                if attempt < max_retries - 1:
                    time.sleep(1 * (2 ** attempt))  # Exponential backoff
                else:
                    print(f"Failed to evaluate after {max_retries} attempts: {e}")
                    return 0.5  # 返回中等分数作为默认值

    def cot_qa_evaluate(self, item):
        """
        保持原有接口兼容性，但使用自定义评估器
        """
        return self.custom_cot_qa_evaluate(item)
        
    def evaluate_with_answer(self, session, answer_gold, task, evaluation_method):
        # return True if failed (score == 0)
        try:
            answer_pred = str(session["steps"][-1]["end"]["final_results"]["output"])
        except:
            answer_pred = "error"
        if evaluation_method == "em":
            _this_corr = int(question_scorer(model_answer=answer_pred, ground_truth=answer_gold))
            return _this_corr == 0
        elif evaluation_method == "llm_score":
            _this_corr = self.cot_qa_evaluate({"pred": answer_pred, "gold": answer_gold, "task": task})
            return _this_corr == 0
        
    def gpt_judge(self, task, pred, traj):
        """
        Processes a single task by extracting the required information, sending it for verification,
        and parsing the response.

        Args:
            data (dict): The task data to be processed.

        Returns:
            dict: A dictionary containing task ID, task description, prediction, gold, verification result, and explanation.
        """
        
        prompt = f"Here is a task description: {task}\n\nHere is an unverified answer: {pred} \n\nHere is the trajectory: {traj} \n\nPlease first provide a concise explanation, then if the unverified answer is correct, return '==yes==', otherwise return '==no=='.\n\n"
        messages = get_messages(prompt, system=get_prompt("gpt_judge_heuristic_with_traj"))

        ans = "yes"
        for i in range(5):
            try:
                ans = self.helper(messages)
                break
            except Exception as e:
                print(f"Error processing task ID {id}: {e}")
                continue

        def parse_answer(ans):
            """Parses the verification response."""
            has_yes = "==yes==" in ans.strip().lower()
            has_no = "==no==" in ans.strip().lower()
            if has_yes and not has_no:
                return "yes"
            elif has_no and not has_yes:
                return "no"
            else:
                return "yes"
        
        explanation = ans.split("<think>")[1].split("</think>")[0].strip() if "<think>" in ans else "na"
        
        return parse_answer(ans), {
            "trajectory": traj,
            "pred": pred,
            "verification": parse_answer(ans),
            "explanation": explanation,
        }
    
    def detect_failure(self, session, evaluation_type):
        failed_to_answer = False
        # final action message
        action_dict = deepcopy(session['steps'][-1]['action'])
        final_message = action_dict['llm_output']
        # end message: output formatting
        end_messages = []
        for i in range(len(session['steps'])):
            if 'end' in session['steps'][i]:
                msg = deepcopy(session['steps'][i]['end']['llm_input'])
                msg.append({"role":"assistant", "content":session['steps'][i]['end']['llm_output']})
                end_messages.append(msg)
        
        if len(end_messages) == 0:
            failed_to_answer = True
        end_message = end_messages[-1][-1]['content']

        if rule_filter_final_action_message(final_message) or rule_filter_end_message(end_message):
            failed_to_answer = True
        
        if evaluation_type == "no_answer":
            return failed_to_answer, "No answer is obtained in the previous try." if failed_to_answer else None
        
        if evaluation_type == "no_answer+no_ask_llm":
            if self.worker_detect_ask_llm(session):
                return True, None
            return failed_to_answer, "Some operations in the previous try failed."
        
        if "gpt_judge" in evaluation_type:
            agent_ans = session["steps"][-1]["end"]["final_results"]["output"]
            agent_traj = self.summarize(session)
            feedback = self.gpt_judge(session["task"], agent_ans, agent_traj)
            if feedback[0] == "no":
                return True, feedback[1]
            return False, None
        else:
            return False, None
        
    def extract_answer_and_log(self, session):
        """Extracts the answer and log from a given instance."""
        ans = session["steps"][-1]["end"]["final_results"]["output"]
        log = self.summarize(session)
        return ans, log
        
    def construct_prompt(self, session_list):
        """Constructs the prompt based on the task and instance list. Optionally reduces log length."""
        task = session_list[0]["task"]
        prompt = "===== Begin of task =====\n\n"
        prompt += task
        for i in range(len(session_list)):
            prompt += f"===== Begin of solution {i} =====\n\n"
            ans, log = self.extract_answer_and_log(session_list[i])
            prompt += f"Answer: {ans}\n\n"
            prompt += f"Log:\n {log}\n\n"
        return prompt
        
    def ensemble(self, session_list):
        prompt = self.construct_prompt(session_list)
        candidates = list(enumerate([x["steps"][-1]["end"]["final_results"]["output"] for x in session_list]))
        try:
            ans = self.helper(get_messages(prompt, system=get_prompt("gpt_chooser"))) 
            print(f"Ensemble answer: {ans}")
        except:
            strings = [item[1] for item in candidates]
            count = Counter(strings)
            # Find the majority string
            majority_string = count.most_common(1)[0][0]
            # Find the index of the majority string
            majority_index = next(i for i, s in candidates if s == majority_string)
            ans = f"<think>Majority voting</think><choice>{majority_index}</choice>"

        if not "<choice>" in ans:
            ans += "<choice>"
        if not "<think>" in ans:
            ans += "<think>"

        return int(ans.split("<choice>")[1].split("</choice>")[0].strip())

        # return {
        #     "choice": int(ans.split("<choice>")[1].split("</choice>")[0].strip()),
        #     "explanation": ans.split("<think>")[1].split("</think>")[0].strip(),
        #     "log": prompt,
        # }
if __name__ == "__main__":
    # for testing
    evaluator = Evaluator()
    task = "What is the capital of France?"
    pred = "Paris"
    traj = "Step 1: Action 1: Search for the capital of France. Observation 1: The capital of France is Paris."
    result = evaluator.gpt_judge(task, pred, traj)
    print(result)

        
