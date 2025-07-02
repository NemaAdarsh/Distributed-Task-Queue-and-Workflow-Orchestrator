<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# FlowCtl - Distributed Task Queue and Workflow Orchestrator

This is a distributed task queue and workflow orchestrator built with Go for the core engine, PostgreSQL for state storage, Redis for task queuing, and React for the web dashboard.

## Code Style Guidelines

- Follow Go best practices and conventions
- Use structured logging with logrus
- Implement proper error handling with context
- Write clean, readable code without excessive comments
- Use dependency injection for testability
- Follow REST API conventions

## Architecture Patterns

- Clean architecture with separated concerns
- Repository pattern for data access
- Factory pattern for component creation
- Observer pattern for event handling
- Command pattern for task execution

## Key Components

- **Scheduler**: Core workflow orchestration engine in Go
- **Workers**: Language-agnostic task executors
- **Storage**: PostgreSQL for persistent state
- **Queue**: Redis for task queuing and coordination
- **API**: REST endpoints for workflow management
- **Dashboard**: React-based web interface

## Development Focus

- Performance and scalability
- Fault tolerance and reliability
- At-least-once delivery guarantees
- Horizontal scaling capabilities
- Real-time monitoring and observability

## Testing Approach

- Unit tests for core business logic
- Integration tests for component interactions
- End-to-end workflow testing
- Performance and load testing
- Chaos engineering practices

When working on this codebase, prioritize:
1. Code clarity and maintainability
2. Performance optimization
3. Error handling and resilience
4. Scalability considerations
5. Security best practices
