# lateral

A simple process parallelizer to make lives better at the commandline.

Have you ever used `xargs -P -n 1` in an attempt to run slow commands in parallel? Gotten burned or gave up? Forgot `-0` when you needed it? Had to use positional arguments? Wanted to substitute text into the commandline?

Lateral is here to help.

Example usage:

     # start a lateral server - one per session (login shell), runs 10 parallel tasks by default
    lateral start
    for i in $(gather list of work); do
      # Hand off the command to lateral
      lateral run -- my_slow_command $i
    done
    # Wait for all the work to be done
    lateral wait
    echo $? # if any commands returned a non-zero status, returns non-zero, also kills the server when complete.

This is the most basic usage, and it's simpler than using xargs.
It also supports much more powerful things:

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
