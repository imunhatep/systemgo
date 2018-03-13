<?php

class TimeRunner
{
    private $callable;
    private $running;
    private $results;

    function __construct(\Generator $runner)
    {
        $this->callable = $runner;
        $this->results = [];
    }

    function start()
    {
        $this->running = true;

        while ($this->running and $this->callable->valid()) {
            /** @var \Generator $round */
            $round = $this->tick();

            foreach ($round as $value) {
                fwrite(STDOUT, $value . "\n");
                //fwrite(STDERR, $value. "\n");
            }
        }
    }

    function stop()
    {
        $this->running = false;
    }

    protected function tick(): \Generator
    {
        foreach ($this->callable as $value) {
            sleep(1);
            yield $value;
        }
    }
}

function counter(int $min, int $max): \Generator
{
    for ($i = $min; $i < $max; $i++) {
        yield $i;
    }
}

$options = getopt("m::x::");
$min = (int) ($options['m'] ?? 0);
$max = (int) ($options['x'] ?? 60);

$runner1 = new TimeRunner(counter($min, $max));
$runner1->start();