# aws-launch

This CLI tool which enables you to launch container based tasks using AWS services (Elastic Container Service (ECS))[https://aws.amazon.com/ecs/] and (CodeBuild)[https://aws.amazon.com/codebuild/]. It provides a simplified interface when using these services, removing some of the inconsistencies.

# Usage

This cli allows you to provide JSON payloads to the API, which are then validated and used to launch and manage tasks. Note there is a dump schema command which will return the JSON schema for each JSON payload.

```
usage: aws-launch [<flags>] <command> [<args> ...]

A command-line task provisioning application.

Flags:
      --help     Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose  Verbose mode.

Commands:
  help [<command>...]
    Show help.


  one-task <one-file>
    Create a new definition and run in one shot.


  define-task <def-file>
    Create a new definition.


  launch-task <launch-file>
    Launch a new task.


  cleanup-task <cleanup-file>
    Cleanup a new task.


  get-task-logs <get-task-logs>
    Get logs for task.


  dump-schema <struct-name>
    Write the JSON Schema to stdout.

```

# License

This project is released under Apache 2.0 License.