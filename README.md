# API LLM Gateway

A robust, intelligent gateway for mapping capabilities to multiple LLM providers, complete with dynamic quota learning and automatic fallbacks.

## Features

- **Format Auto-Detection Middleware**: Automatically routes and translates OpenAI, Anthropic, and other formats transparently.
- **Universal Client Compatibility**: Out-of-the-box support for modern AI clients (OpenWebUI, Claude Code, OpenHands, UI-Tars, CrewAI).
- **Dynamic Quota Learning**: Automatically extracts `Remaining` and `ResetAt` from Provider HTTP headers.
  - Persists learned quotas to PostgreSQL so they survive reboots.
  - Penalizes router scores dynamically when a provider's quota is below 20%.
  - Recognizes `Retry-After` headers during 429 Too Many Requests to cleanly retire failing models.
- **Failover & Routing**: Routes requests based on dynamic scores (Quality, Latency, Cost, Quota, Health).

## Setup

Read the guides in `docs/14-guides/` for setting up your preferred AI clients with this Gateway.
