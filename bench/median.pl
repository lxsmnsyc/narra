#!/usr/bin/perl
# Usage: median.pl RUNS EXECUTABLE
# Runs the executable RUNS times. Prints "<median seconds> <stdout of the last run>".
use strict;
use warnings;
use Time::HiRes qw(time);

my ($runs, $exe) = @ARGV;
my @times;
my $out = "";
for (1 .. $runs) {
    my $start = time();
    open(my $fh, "-|", $exe) or die "cannot run $exe: $!";
    $out = do { local $/; <$fh> } // "";
    close($fh);
    my $status = $?;
    push @times, time() - $start;
    if ($status != 0) {
        print "FAIL exit=" . ($status >> 8) . "\n";
        exit 1;
    }
}
@times = sort { $a <=> $b } @times;
my $mid = int(@times / 2);
my $median = @times % 2 ? $times[$mid] : ($times[$mid - 1] + $times[$mid]) / 2;
$out =~ s/\s+$//;
printf "%.3f %s\n", $median, $out;
