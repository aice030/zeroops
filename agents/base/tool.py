#

import requests
from .utils import KwargsInitializable, rprint, GET_ENV_VAR

class Tool(KwargsInitializable):
    def __init__(self, **kwargs):
        self.name = ""
        super().__init__(**kwargs)

    def get_function_definition(self, short: bool):
        raise NotImplementedError("To be implemented")

    def __call__(self, *args, **kwargs):
        raise NotImplementedError("To be implemented")

# --
# useful tools

class StopResult(dict):
    pass

class StopTool(Tool):
    def __init__(self, agent=None):
        super().__init__(name="stop")
        self.agent = agent

    def get_function_definition(self, short: bool):
        if short:
            return """- def stop(output: str, log: str) -> Dict:  # Finalize and formalize the answer when the task is complete."""
        else:
            return """- stop
```python
def stop(output: str, log: str) -> dict:
    \""" Finalize and formalize the answer when the task is complete.
    Args:
        output (str): The concise, well-formatted final answer to the task.
        log (str): Brief notes or reasoning about how the answer was determined.
    Returns:
        dict: A dictionary with the following structure:
            {
                'output': <str>  # The well-formatted answer, strictly following any specified output format.
                'log': <str>     # Additional notes, such as steps taken, issues encountered, or relevant context.
            }
    Examples:
        >>> answer = stop(output="Inter Miami", log="Task completed. The answer was found using official team sources.")
        >>> print(answer)
    \"""
```"""

    def __call__(self, output: str, log: str):
        ret = StopResult(output=output, log=log)
        if self.agent is not None:
            self.agent.put_final_result(ret)  # mark end and put final result
        return ret

class AskLLMTool(Tool):
    def __init__(self, llm=None):
        super().__init__(name="ask_llm")
        self.llm = llm

    def set_llm(self, llm):
        self.llm = llm

    def get_function_definition(self, short: bool):
        if short:
            return """- def ask_llm(query: str) -> str:  # Directly query the language model for tasks that do not require external tools."""
        else:
            return """- ask_llm
```python
def ask_llm(query: str) -> str:
    \""" Directly query the language model for tasks that do not require external tools.
    Args:
        query (str): The specific question or instruction for the LLM.
    Returns:
        str: The LLM's generated response.
    Notes:
        - Use this function for fact-based or reasoning tasks that can be answered without web search or external data.
        - Phrase the query clearly and specifically.
    Examples:
        >>> answer = ask_llm(query="What is the capital city of the USA?")
        >>> print(answer)
    \"""
```"""

    def __call__(self, query: str):
        messages = [{"role": "system", "content": "You are a helpful assistant. Answer the user's query with your internal knowledge. Ensure to follow the required output format if specified."}, {"role": "user", "content": query}]
        response = self.llm(messages)
        return response

class SimpleSearchTool(Tool):
    def __init__(self, target="", llm=None, max_results=7, list_enum=True, **kwargs):
        super().__init__(name="simple_web_search")
        self.llm = llm
        self.max_results = max_results
        self.list_enum = list_enum
        if not target:
            target = GET_ENV_VAR("SEARCH_BACKEND", df="DuckDuckGo")  # use which backend search engine
        rprint(f"Setup SimpleSearchTool with {target}")
        self.target = target
        if target == "DuckDuckGo":
            self.ddgs_params = kwargs.copy()
        elif target == "Google":
            self.google_params = {"key": GET_ENV_VAR("SEARCH_API_KEY"), "cx": GET_ENV_VAR("SEARCH_CSE_ID")}
        else:
            raise ValueError(f"UNK search target = {target}")
        # --

    def set_llm(self, llm):
        self.llm = llm  # might be useful for formatting?

    def get_function_definition(self, short: bool):
            if short:
                return """- def simple_web_search(query: str) -> str:  # Perform a quick web search using a search engine for straightforward information needs."""
            else:
                return """- simple_web_search
```python
def simple_web_search(query: str) -> str:
    \""" Perform a quick web search using a search engine for straightforward information needs.
    Args:
        query (str): A simple, well-phrased search term or question.
    Returns:
        str: A string containing search results, including titles, URLs, and snippets.
    Notes:
        - Use for quick lookups or when you need up-to-date information.
        - Avoid complex or multi-step queries; keep the query simple and direct.
        - Do not use for tasks requiring deep reasoning or multi-source synthesis.
    Examples:
        >>> answer = simple_web_search(query="latest iPhone")
        >>> print(answer)
    \"""
```"""

    def __call__(self, query: str):
        target = self.target
        if target == "DuckDuckGo":
            from duckduckgo_search import DDGS
            ddgs = DDGS(**self.ddgs_params)
            rprint(f"Query ddgs with: query={query}, max_results={self.max_results}")
            results = ddgs.text(query, max_results=self.max_results)
            search_results = [{"title": _item["title"], "link": _item["href"], "content": _item["body"]} for _item in results]
        elif target == "Google":
            url = "https://www.googleapis.com/customsearch/v1"
            params = self.google_params.copy()
            params.update({"q": query, "num": self.max_results})
            rprint(f"Query google-search with params={params}")
            response = requests.get(url, params=params)
            results = response.json()
            search_results = [{"title": _item["title"], "link": _item["link"], "content": _item["snippet"]} for _item in results.get("items", [])]
        else:
            raise ValueError(f"UNK search target = {target}")
        # --
        if len(search_results) == 0:
            ret = "Search Results: No results found! Try a less restrictive/simpler query."
        elif self.list_enum:
            ret = "Search Results:\n" + "\n".join([f"({ii}) title={repr(vv['title'])}, link={repr(vv['link'])}, content={repr(vv['content'])}" for ii, vv in enumerate(search_results)])
        else:
            ret = "Search Results:\n" + "\n".join([f"- title={repr(vv['title'])}, link={repr(vv['link'])}, content={repr(vv['content'])}" for ii, vv in enumerate(search_results)])
        return ret

class QueryErrorLogsTool(Tool):
    def __init__(self, llm=None):
        super().__init__(name="query_error_logs")
        self.llm = llm

    def set_llm(self, llm):
        self.llm = llm

    def get_function_definition(self, short: bool):
        if short:
            return """- def query_error_logs(host: str, start_time: str, end_time: str) -> list::  # Queries and retrieves IMPORTANT logs (ERROR, WARN, FATAL) for a specific host within a time range."""
        else:
            return """- query_error_logs
    ```python
    def query_error_logs(host: str, start_time: str, end_time: str) -> list:
        \"""
        Queries and retrieves IMPORTANT logs (ERROR, WARN, FATAL) for a specific host within a given time range.

        Args:
            host (str): The target hostname/IP to query logs from
            start_time (str): Start time of the query range (ISO 8601 format)
            end_time (str): End time of the query range (ISO 8601 format)

        Returns:
            list: An array of raw IMPORTANT log strings (levels: ERROR, WARN, FATAL), empty if none found.
            
            Example outputs:
            [
                '2025-08-15T14:10:27Z WARN  [GoroutineLeakDetector] Detected goroutine leak: 15432 goroutines active (expected < 500)',
                '2025-08-15T14:12:33Z FATAL [Runtime] fatal error: runtime: out of memory',
                '2025-08-15T14:12:33Z ERROR [Main] request failed after retries: context deadline exceeded'
            ]
            
            or when no important logs:
            []

        Notes:
            1. Time format must be ISO 8601 compliant (YYYY-MM-DDThh:mm:ssZ).
            2. Returns only levels in {"ERROR", "WARN", "FATAL"}; INFO/DEBUG/TRACE are excluded.
            3. Logs are returned in chronological order (oldest first).
            4. The complete raw log line is preserved including timestamps.
            5. Time range is inclusive (logs exactly at start/end time are included).
            6. Maximum query range is 30 days (returns error if exceeded).
            7. Host must exist in the monitoring system.
            8. Returns empty array [] when no matching logs found.
            9. When multiple lines share the same timestamp, the original source order is preserved if available.

        Examples:
            >>> query_error_logs(
            ...     'web-server-01',
            ...     '2025-08-15T00:00:00Z',
            ...     '2025-08-15T23:59:59Z'
            ... )
            [
                '2025-08-15T03:45:22Z WARN  [nginx] upstream server temporarily disabled for 30s',
                '2025-08-15T14:10:27Z WARN  [GoroutineLeakDetector] Detected goroutine leak: 15432 goroutines active (expected < 500)',
                '2025-08-15T14:12:33Z FATAL [Runtime] fatal error: runtime: out of memory'
            ]

            >>> query_error_logs(
            ...     'db-server-01',
            ...     '2025-08-01T00:00:00Z',
            ...     '2025-08-31T00:00:00Z'
            ... )
            []  # No important logs during this period
        \"""
    ```"""


    def __call__(self, host: str, start_time: str, end_time: str):
        """
        Implementation of the query_error_logs tool.
        This is a mock implementation that returns sample error logs.
        In a real implementation, this would connect to a logging system.
        """
        import re
        from datetime import datetime, timedelta
        
        # Validate input parameters
        if not host or not isinstance(host, str):
            raise ValueError("Host must be a non-empty string")
        
        if not start_time or not isinstance(start_time, str):
            raise ValueError("Start time must be a non-empty string")
            
        if not end_time or not isinstance(end_time, str):
            raise ValueError("End time must be a non-empty string")
        
        # Validate ISO 8601 time format
        try:
            start_dt = datetime.fromisoformat(start_time.replace('Z', '+00:00'))
            end_dt = datetime.fromisoformat(end_time.replace('Z', '+00:00'))
        except ValueError:
            raise ValueError("Time format must be ISO 8601 compliant (YYYY-MM-DDThh:mm:ssZ)")
        
        # Check if time range exceeds 30 days
        if (end_dt - start_dt).days > 30:
            raise ValueError("Maximum query range is 30 days")
        
        # Check if start time is before end time
        if start_dt >= end_dt:
            raise ValueError("Start time must be before end time")
        
        # Mock implementation - in real scenario, this would query actual log systems
        # like Elasticsearch, Splunk, or other logging platforms
        mock_logs = [
            "2025-08-15T14:10:27Z WARN  [GoroutineLeakDetector] Detected goroutine leak: 15432 goroutines active (expected < 500)",
            "2025-08-15T14:10:27Z WARN  [GoroutineLeakDetector] Sample leaked goroutine stack:goroutine 112233 [IO wait]:net.(*conn).Read(0xc000ab1230, 0xc0012c0000, 4096, 4096, 0x0, 0x0, 0x0)/usr/local/go/src/net/net.go:184io.copyBuffer(0x7f98f3c2d0, 0xc000a4f500, 0x7f98f3c2a0, 0xc001c8c000, 0xc0012c0000, 0x1000, 0x2000, 0x0, 0x0, 0x0)/usr/local/go/src/io/io.go:422myservice/stream.(*Handler).StartStream.func1()/app/stream/handler.go:85created by myservice/stream.(*Handler).StartStream/app/stream/handler.go:72",
            "2025-08-15T14:12:33Z FATAL [Runtime] fatal error: runtime: out of memory"
        ]
        
        return mock_logs
    
class QueryDependencyTool(Tool):
    def __init__(self, llm=None):
        super().__init__(name="query_dependency")
        self.llm = llm
    
    def set_llm(self, llm):
        self.llm = llm

    def get_function_definition(self, short: bool):
        if short:
            return "- def query_dependency(target_service: str) -> list[list[str]]:  # Finds complete upstream-downstream call chains for a target service and returns all possible paths as a nested array."
        else:
            return """- query_dependency
```python
def query_dependency(target_service: str) -> list[list[str]]:

    \"""
    Finds complete upstream-downstream call chains for a target service and returns all possible paths as a nested array.

    Args:
        target_service: The service name to query (e.g., 'C' in the example)

    Returns:
        A nested list where each sublist represents a complete call chain from the most upstream 
        to the most downstream service (e.g., [['A','B','C','D','E'], ['A','B','C','F','G']])

    Example:
        >>> find_service_relation_chains('C')
        [['A', 'B', 'C', 'D', 'E'], ['A', 'B', 'C', 'F', 'G']]

    Notes:
        1. The returned chains include the target service itself (e.g., 'C' in the example)
        2. Each chain represents a complete end-to-end path (from root service to terminal service)
        3. Returns empty list if no related chains are found
        4. Service names are case-sensitive
        5. The order within each chain reflects actual invocation sequence
        6. May return multiple independent chains when bifurcations exist downstream
    Examples:
        >>> chains = find_service_relation_chains('C')
        >>> print(chains)  # Output shows all possible call chains through service C
        [['A', 'B', 'C', 'D', 'E'], ['A', 'B', 'C', 'F', 'G']]
        
        >>> chains = find_service_relation_chains('B')
        >>> print(chains)  # Output shows all chains through service B (including bifurcations)
        [['A', 'B', 'C', 'D', 'E'], ['A', 'B', 'C', 'F', 'G'], ['X', 'B', 'Y']]
    \"""
```
            """
        
    def __call__(self, target_service: str):
        """
        Implementation of the query_dependency tool.
        This is a mock implementation that returns sample dependency chains.
        In a real implementation, this would query a service dependency graph.
        """
        # Validate input parameters
        if not target_service or not isinstance(target_service, str):
            raise ValueError("Target service must be a non-empty string")
        
        # Mock implementation - in real scenario, this would query actual service dependency systems
        # like Jaeger, Zipkin, or other distributed tracing platforms
        
        # Sample dependency chains for different services
        mock_dependencies = {
             "C": [
                 ["A", "B", "C", "D", "E"],
                 ["A", "B", "C", "F", "G"]
             ],
             "B": [
                 ["A", "B", "C", "D", "E"],
                 ["A", "B", "C", "F", "G"],
                 ["X", "B", "Y"]
             ],
             "A": [
                 ["A", "B", "C", "D", "E"],
                 ["A", "B", "C", "F", "G"],
                 ["A", "H", "I"]
             ],
             "D": [
                 ["A", "B", "C", "D", "E"]
             ],
             "E": [
                 ["A", "B", "C", "D", "E"]
             ],
             "F": [
                 ["A", "B", "C", "F", "G"]
             ],
             "G": [
                 ["A", "B", "C", "F", "G"]
             ],
             "H": [
                 ["A", "H", "I"]
             ],
             "I": [
                 ["A", "H", "I"]
             ],
             "X": [
                 ["X", "B", "Y"]
             ],
             "Y": [
                 ["X", "B", "Y"]
             ]
         }
        
        # Return dependency chains for the target service, or empty list if not found
        return mock_dependencies.get(target_service, [])