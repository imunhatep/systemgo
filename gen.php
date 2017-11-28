<?php

class TimeRunner
{
    private $callable;
    private $stop;
    private $results;

    function __construct(\Generator $runner)
    {
        $this->callable = $runner;
        $this->results = [];
    }

    function start()
    {
        $this->stop = false;

        while (!$this->stop and $this->callable->valid()) {
            /** @var \Generator $round */
            $round = $this->tick();

            foreach ($round as $value) {
                $this->results[] = $value;

                fwrite(STDOUT, $value . "\n");
                //fwrite(STDERR, $value. "\n");

				error_log(sprintf("[%s] Running: %s\n", date("H:i:s"), $value), 3, "/tmp/gen.log");
            }
        }
    }

    function stop()
    {
        $this->stop = true;
    }

    function flush(): array
    {
        $results = $this->results and $this->results = [];

        return $results;
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