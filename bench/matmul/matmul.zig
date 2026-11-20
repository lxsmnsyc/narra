const std = @import("std");

fn multiply(allocator: std.mem.Allocator, a: []const i64, b: []const i64, n: usize) ![]i64 {
    const c = try allocator.alloc(i64, n * n);
    @memset(c, 0);
    for (0..n) |i| {
        for (0..n) |j| {
            var sum: i64 = 0;
            for (0..n) |k| {
                sum += a[i * n + k] * b[k * n + j];
            }
            c[i * n + j] = sum;
        }
    }
    return c;
}

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    var n: usize = 600;
    std.mem.doNotOptimizeAway(&n);
    const a = try allocator.alloc(i64, n * n);
    defer allocator.free(a);
    const b = try allocator.alloc(i64, n * n);
    defer allocator.free(b);
    for (0..n) |i| {
        for (0..n) |j| {
            a[i * n + j] = @intCast((i + j) % 10);
            b[i * n + j] = @intCast((i * j) % 10);
        }
    }
    const c = try multiply(allocator, a, b, n);
    defer allocator.free(c);
    var total: i64 = 0;
    for (c) |x| {
        total += x;
    }
    var buffer: [64]u8 = undefined;
    var writer = std.Io.File.stdout().writer(init.io, &buffer);
    try writer.interface.print("{d}\n", .{total});
    try writer.interface.flush();
}
