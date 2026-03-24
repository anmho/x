FROM oven/bun:1.3.2

WORKDIR /app

COPY packages/sdk-agent-control /app/packages/sdk-agent-control
COPY apps/agent-runner /app/apps/agent-runner

WORKDIR /app/apps/agent-runner

RUN bun install

# Claude SDK still depends on the Claude runtime being available in the image.
RUN bun install -g @anthropic-ai/claude-code

ENV HOME=/tmp/agent-home
RUN mkdir -p /tmp/agent-home/.claude

ENTRYPOINT ["bun", "src/index.ts"]
