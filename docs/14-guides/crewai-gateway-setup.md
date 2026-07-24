# CrewAI con API LLM Gateway

## Configuración de Python SDK

CrewAI utiliza Langchain/OpenAI por debajo. Puedes conectarlo al Gateway con unas pocas líneas de código.

```python
import os
from crewai import Agent, Task, Crew
from langchain_openai import ChatOpenAI

os.environ["OPENAI_API_BASE"] = "http://localhost:8080/v1"
os.environ["OPENAI_API_KEY"] = "tu_api_key_gateway"

# Usar el Gateway para decidir el modelo
llm = ChatOpenAI(model_name="router:capability:agents")

researcher = Agent(
    role='Senior Researcher',
    goal='Discover new technologies',
    backstory='You are a curious mind',
    llm=llm
)
```
