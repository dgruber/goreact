# web

This example contains 4 commands:
- search: allows the LLM to perform a Google search
- bing_search: allows the LLM to perform a Bing search
- scrape: allows the LLM to fetch any web page and get its content
- ask: allows the LLM to include the user in its loop for further questions

## Example run

QUESTION: What is the answer to life, the universe and everything?
THOUGHT: This seems like a philosophical question. I'm not sure what the answer is, but I'll try to search for it.
ACTION: { "command": "search", "args": "answer to life universe everything" } 
{ "command": "search", "args": "answer to life universe everything" } 
EXECUTING COMMAND: search answer to life universe everything
Search result: [{1 https://news.mit.edu/2019/answer-life-universe-and-everything-sum-three-cubes-mathematics-0910 The answer to life, the universe, and everything | MIT News 10 Sept 2019 —} {2 https://en.wikipedia.org/wiki/Phrases_from_The_Hitchhiker%27s_Guide_to_the_Galaxy Phrases from The Hitchhiker's Guide to the Galaxy42 (number) - Wikipedia This leaves the symbolic meaning that the answer to life, the universe, and everything is anything you, the user, would like it to be. 42 PuzzleEdit. The 42 ...The number 42 is, in The Hitchhiker's Guide to the Galaxy by Douglas Adams, the "Answer to the Ultimate Question of Life, the Universe, and Everything," ...} {3 https://www.scientificamerican.com/article/for-math-fans-a-hitchhikers-guide-to-the-number-42/ For Math Fans: A Hitchhiker's Guide to the Number 42 21 Sept 2020 —}]
OBSERVATION:  The answer to the question "What is the answer to life, the universe and everything?" according to Douglas Adams' book "The Hitchhiker's Guide to the Galaxy" is the number 42. This answer is meant to be symbolic and open to interpretation. In mathematics, there is a puzzle called the "42 Puzzle". However, there is no real answer to the question in a literal sense.
THOUGHT: According to my search, there is no literal answer to the question. However, the number 42 appears to hold some significance. Should I ask the user if they are satisfied with this information or if they want me to look for something else? 
EXECUTING COMMAND: ask Are you satisfied with this information or would you like me to look for something else?
Please answer the question: Are you satisfied with this information or would you like me to look for something else?
It is ok.
OBSERVATION:  It is ok.

CONCLUSION: Since the user responded with "It is ok", I can conclude that they are satisfied with the information. Thus, I can provide the answer to them based on my observation.
ANSWER: According to Douglas Adams' book "The Hitchhiker's Guide to the Galaxy", the answer to the question "What is the answer to life, the universe and everything?" is the number 42 which is meant to be symbolic and open to interpretation. However, there is no real answer to the question in a literal sense.

## Bing Search Example

QUESTION: What is the capital of France?
THOUGHT: I need to find the capital of France. I will use the Bing search engine to find the answer.
ACTION: { "command": "bing_search", "args": "capital of France" }
EXECUTING COMMAND: bing_search capital of France
Bing search result: [{1 https://en.wikipedia.org/wiki/Paris Paris - Wikipedia Paris is the capital and most populous city of France. It has an area of 105 square kilometers and a population of 2,148,271 in 2020.}]
OBSERVATION: The capital of France is Paris.
THOUGHT: I have found the answer to the question.
ANSWER: The capital of France is Paris.
