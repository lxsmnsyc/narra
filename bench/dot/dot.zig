const std = @import("std");

// Eight scalar accumulators, summed in the order of Simd.sum in dot.nra.
fn dot(left: []const f32, right: []const f32, passes: usize) f32 {
    var s: [8]f32 = @splat(0.0);
    for (0..passes) |_| {
        var i: usize = 0;
        while (i + 8 <= left.len and i + 8 <= right.len) : (i += 8) {
            for (0..8) |k| {
                s[k] += left[i + k] * right[i + k];
            }
        }
    }
    return ((s[0] + s[4]) + (s[2] + s[6])) + ((s[1] + s[5]) + (s[3] + s[7]));
}

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const count: usize = 1048576;
    const left = try allocator.alloc(f32, count);
    defer allocator.free(left);
    const right = try allocator.alloc(f32, count);
    defer allocator.free(right);
    for (0..count) |index| {
        const i: i64 = @intCast(index);
        left[index] = @as(f32, @floatFromInt(@mod(i, 16) - 8)) * 0.125;
        right[index] = @as(f32, @floatFromInt(@mod(i, 7) - 3)) * 0.25;
    }
    var passes: usize = 200;
    std.mem.doNotOptimizeAway(&passes);
    const result = dot(left, right, passes);
    var buffer: [64]u8 = undefined;
    var writer = std.Io.File.stdout().writer(init.io, &buffer);
    try writer.interface.print("{d}\n", .{@as(i64, @intFromFloat(result * 32.0))});
    try writer.interface.flush();
}
