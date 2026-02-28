FROM ollama/ollama:latest

ARG MODEL=llama3.2:1b

RUN bash -c "ollama serve &" && \
    until ollama list > /dev/null 2>&1; do sleep 1; done && \
    ollama pull $MODEL
