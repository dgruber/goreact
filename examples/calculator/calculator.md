# Calculator example of using ReAct in Go

Here is a sample run:

QUESTION: What is the square root of 10? What is PI? What is the sum of both numbers?
THOUGHT: I need to calculate the square root of 10 and find the value of PI.
ACTION: { "command": "calculate", "args": "sqrt(10)" } 
OBSERVATION: 3.162278
THOUGHT: I have the square root of 10, now I need to find the value of PI.
ACTION: { "command": "calculate", "args": "pi" } 
OBSERVATION: 3.141593
THOUGHT: I have the value of PI, now I need to calculate the sum of both numbers.
ACTION: { "command": "calculate", "args": "3.162278+3.141593" } 
OBSERVATION: 6.303871
Answer: I have the sum of the square root of 10 and PI.
ANSWER: The sum of the square root of 10 and PI is 6.303871.