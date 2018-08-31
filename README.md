# lateral

A simple process parallelizer to make lives better at the commandline.

Have you ever used `xargs -P -n 1` in an attempt to run slow commands in parallel? Gotten burned or gave up? Forgot `-0` when you needed it? Had to use positional arguments? Wanted to substitute text into the commandline?

Lateral is here to help.

## Usage

    Usage:
      lateral [command]
 
    Available Commands:
      config      Change the server configuration
      dumpconfig  Dump available configuration options
      getpid      Print pid of server to stdout
      kill        Kill the server with fire
      run         Run the given command in the lateral server
      start       Start the lateral background server
      wait        Wait for all currently inserted tasks to finish
 
    Flags:
          --config string   config file (default $HOME/.lateral/config.yaml)
      -h, --help            help for lateral
      -s, --socket string   UNIX domain socket path (default $HOME/.lateral/socket.$SESSIONID)

## Examples

Example usage:

    lateral start
    for i in $(cat /tmp/names); do
      lateral run -- some_command $i
    done
    lateral wait
    
With comments:

     # start a lateral server - one per session (login shell), runs 10 parallel tasks by default
    lateral start
    for i in $(gather list of work); do
      # Hand off the command to lateral
      lateral run -- my_slow_command $i
    done
    # Wait for all the work to be done
    lateral wait
    # wait stops the server when all commands are complete
    # if any commands returned a non-zero status, wait returns non-zero
    echo $?

This is the most basic usage, and it's simpler than using xargs.
It also supports much more powerful things. It doesn't only support command-line arguments
like xargs, it also supports filedescriptors:

    lateral start
    for i in $(seq 1 100); do
      lateral run -- my_slow_command < workfile$i > /tmp/logfile$i
    done
    lateral wait

The stdin, stdout, and stderr of the command to be run are passed to lateral, and so redirection to files works. This makes it trivial to have per-task log files.

The parallelism is also dynamically adjustable at run-time.

    lateral start -p 0 # yup, it will just queue tasks with 0 parallelism
    for i in $(seq 1 100); do
      lateral run -- command_still_outputs_to_stdout_but_wont_spam inputfile$i
    done
    lateral config -p 10; lateral wait # command output spam can commence


This also allows you to raise parallelism when things are going slower than you want. Underestimate how much work your machine can do at once? Ratchet up the number of tasks with `lateral config -p <N>`.
Turns out that you want to run fewer? Reducing the parallelism works as well - no new tasks will be started until the number running is under the limit.

## How does this black magic work?

`lateral start` starts a server that listens on a unix socket that (by default) is based on your session ID. Each invocation of `lateral` in a different login shell will be independent.

Unix sockets support more than passing data: they also support passing filedescriptors. When `lateral run` is called, it sends the command-line, environment, and open filedescriptors over the unix socket to the server where they are stored and queued to be run.

## Logging

Lateral uses glog, a Google logging library. Only the server logs, other commands do not. Its logs by default can be found in /tmp/lateral.INFO, but this can be configured using standard Google logging flags.
