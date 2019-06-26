#!/usr/bin/php
<?php
/**
 * Monitor Beanstalkd queues health and return queues that are not being processed.
 * @author mdroog
 */
$options = getopt("v::");
define("VERBOSE", isset($options["v"]));

function res_kvp($conn, $cmd) {
    if (VERBOSE) echo ">> $cmd\n";
    if (fwrite($conn, $cmd."\r\n", strlen($cmd."\r\n")) === false) {
        echo "UNKNOWN - fwrite($cmd) fail\n";
        exit(3);
    }

    $bytes = 0;
    {
        $res = stream_get_line($conn, 16384, "\r\n");
        if (VERBOSE) echo "<< $res\n";
        if ($res === false || substr($res, 0, 3) !== "OK ") {
            echo "CRITICAL - reply($cmd) wrong: $res\n";
            exit(2);
        }
        $bytes = substr($res, 3);
    }

    $lines = array();
    {
    $res = trim(fread($conn, $bytes+2));
    if ($res === false) {
        echo "CRITICAL - reply($cmd) wrong: $res\n";
        exit(2);
    }
    if (VERBOSE) echo "<< $res\n";
    foreach (explode("\n", $res) as $line) {
        if (strlen($line) === 0) {
            // Empty line
            continue;
        }
        if ($line === "---") {
            // Skip separation-thingy
            continue;
        }
        $sep = strpos($line, ":");
        if ($sep === false) {
            echo "CRITICAL - expected separation char on line: $line\n";
            exit(2);
        }
        $lines[ substr($line, 0, $sep) ] = substr($line, $sep+2);

    }
    }
    return $lines;
}
function res_list($conn, $cmd) {
    if (VERBOSE) echo ">> $cmd\n";
    if (fwrite($conn, $cmd."\r\n", strlen($cmd."\r\n")) === false) {
        echo "UNKNOWN - fwrite($cmd) fail\n";
        exit(3);
    }

    $bytes = 0;
    {
        $res = stream_get_line($conn, 16384, "\r\n");
        if (VERBOSE) echo "<< $res\n";
        if ($res === false || substr($res, 0, 3) !== "OK ") {
            echo "CRITICAL - reply($cmd) wrong: $res\n";
            exit(2);
        }
        $bytes = substr($res, 3);
    }

    $lines = array();
    {
    $res = trim(fread($conn, $bytes+2));
    if ($res === false) {
        echo "CRITICAL - reply($cmd) wrong: $res\n";
        exit(2);
    }
    foreach (explode("\n", $res) as $line) {
        if (VERBOSE) echo "<< $line\n";
        if (strlen($line) === 0) {
            // Empty line
            continue;
        }
        if ($line === "---") {
            // Skip separation-thingy
            continue;
        }
        $sep = strpos($line, "- ");
        if ($sep === false) {
            echo "CRITICAL - expected separation char on line: $line\n";
            exit(2);
        }
        $lines[] = substr($line, $sep+2);

    }
    }
    return $lines;
}

$errno = null;
$errstr = null;
//$conn = stream_socket_client("tcp://127.0.0.1:11300", $errno, $errstr, 2);
$conn = fsockopen("127.0.0.1", 11300, $errno, $errstr, 2);
if ($conn === false) {
    echo "CRITICAL - stream_socket_client fail: $errno $errstr\n";
    exit(2);
}
if (stream_set_timeout($conn, 30) === false) {
    echo "UNKNOWN - stream_set_timeout fail\n";
    exit(3);
}
$tubes = res_list($conn, "list-tubes");

foreach ($tubes as $tube) {
    $lines = res_kvp($conn, "stats-tube $tube");
    if ($lines["current-jobs-ready"] > 10) {
        echo sprintf("CRITICAL - queue(%s) current-jobs-ready(%d)\n", $tube, $lines["current-jobs-ready"]);
        exit(2);
    } else if ($lines["current-jobs-ready"] > 5) {
        echo sprintf("WARNING - queue(%s) current-jobs-ready(%d)\n", $tube, $lines["current-jobs-ready"]);
        exit(1);
    }

    if ($lines["current-jobs-buried"] > 10) {
        echo sprintf("CRITICAL - queue(%s) current-jobs-buried(%d)\n", $tube, $lines["current-jobs-buried"]);
        exit(2);
    } else if ($lines["current-jobs-buried"] > 0) {
        echo sprintf("WARNING - queue(%s) current-jobs-buried(%d)\n", $tube, $lines["current-jobs-buried"]);
        exit(1);
    }
}

if (fwrite($conn, "quit") === false) {
    echo "UNKNOWN - fwrite fail?\n";
    exit(3);
}
fclose($conn);

echo "OK";
exit(0);
