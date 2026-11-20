const std = @import("std");

fn countPrimes(allocator: std.mem.Allocator, n: usize) !i64 {
    const marks = try allocator.alloc(i64, n);
    defer allocator.free(marks);
    @memset(marks, 1);
    marks[0] = 0;
    marks[1] = 0;
    var i: usize = 2;
    while (i * i < n) {
        if (marks[i] == 1) {
            var j = i * i;
            while (j < n) {
                marks[j] = 0;
                j += i;
            }
        }
        i += 1;
    }
    var count: i64 = 0;
    for (0..n) |k| {
        count += marks[k];
    }
    return count;
}

pub fn main(init: std.process.Init) !void {
    var n: usize = 20000000;
    std.mem.doNotOptimizeAway(&n);
    var buffer: [64]u8 = undefined;
    var writer = std.Io.File.stdout().writer(init.io, &buffer);
    try writer.interface.print("{d}\n", .{try countPrimes(init.gpa, n)});
    try writer.interface.flush();
}
