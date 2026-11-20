const std = @import("std");

const Op = union(enum) {
    add: i64,
    sub: i64,
    mul: i64,
    mod: i64,
    nop,
};

fn step(op: Op, acc: i64) i64 {
    return switch (op) {
        .add => |n| acc + n,
        .sub => |n| acc - n,
        .mul => |n| acc * n,
        .mod => |n| @rem(acc, n),
        .nop => acc,
    };
}

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    var program: std.ArrayList(Op) = .empty;
    defer program.deinit(allocator);
    var i: i64 = 0;
    while (i < 1000) : (i += 1) {
        try program.append(allocator, switch (@mod(i, 5)) {
            0 => Op{ .add = i },
            1 => Op{ .mul = 3 },
            2 => Op{ .sub = 7 },
            3 => Op{ .mod = 1000003 },
            else => Op.nop,
        });
    }
    var acc: i64 = 1;
    var rounds: i64 = 150000;
    std.mem.doNotOptimizeAway(&rounds);
    var round: i64 = 0;
    while (round < rounds) : (round += 1) {
        for (program.items) |op| {
            acc = step(op, acc);
        }
    }
    var buffer: [64]u8 = undefined;
    var writer = std.Io.File.stdout().writer(init.io, &buffer);
    try writer.interface.print("{d}\n", .{acc});
    try writer.interface.flush();
}
