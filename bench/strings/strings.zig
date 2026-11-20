const std = @import("std");

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    var s: []u8 = try allocator.alloc(u8, 0);
    var count: i64 = 100000;
    std.mem.doNotOptimizeAway(&count);
    var i: i64 = 0;
    while (i < count) : (i += 1) {
        const next = try std.fmt.allocPrint(allocator, "{s}{d}", .{ s, @mod(i, 10) });
        allocator.free(s);
        s = next;
    }
    var buffer: [64]u8 = undefined;
    var writer = std.Io.File.stdout().writer(init.io, &buffer);
    try writer.interface.print("{d}\n", .{s.len});
    try writer.interface.flush();
    allocator.free(s);
}
