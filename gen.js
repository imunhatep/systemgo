#!/usr/bin/env node
'use strict';

/** @var Map options **/
const options = new Map

// print process.argv
process.argv.forEach(
    (val, index, array) => {
        if(index < 2) return

        let [key, value] = val.split('=')

        if(key !== 'undefined' && value !== 'undefined'){
            options.set(key.replace('-',''), value)
        }
    }
);

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
}


const min = parseInt(options.has('m') ? options.get('m') : 0)
const max = parseInt(options.has('x') ? options.get('x') : 60)


const range = (m, x) => [...Array(x).keys()].slice(m)

range(min, max).forEach(
    (v, k) => {
        sleep(1000 * k).then(
            () => process.stdout.write(v.toString() + "\n")
        )
    }
)

