const std = @import("std");

fn fib(n: i64) i64 {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

pub fn main(init: std.process.Init) !void {
    var n: i64 = 38;
    std.mem.doNotOptimizeAway(&n);
    var buffer: [64]u8 = undefined;
    var writer = std.Io.File.stdout().writer(init.io, &buffer);
    try writer.interface.print("{d}\n", .{fib(n)});
    try writer.interface.flush();
}
