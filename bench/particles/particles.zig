const std = @import("std");

const Particle = struct {
    x: i64,
    y: i64,
    vx: i64,
    vy: i64,
};

fn advance(p: Particle) Particle {
    const nx = p.x + p.vx;
    const ny = p.y + p.vy;
    var nvx = p.vx;
    var nvy = p.vy;
    if (nx < 0 or nx > 10000) {
        nvx = -nvx;
    }
    if (ny < 0 or ny > 10000) {
        nvy = -nvy;
    }
    return Particle{ .x = nx, .y = ny, .vx = nvx, .vy = nvy };
}

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    var particles: std.ArrayList(Particle) = .empty;
    defer particles.deinit(allocator);
    var i: i64 = 0;
    while (i < 1000) : (i += 1) {
        try particles.append(allocator, Particle{ .x = @mod(i * 7, 10000), .y = @mod(i * 13, 10000), .vx = @mod(i, 11) + 1, .vy = @mod(i, 7) + 1 });
    }
    var rounds: i64 = 100000;
    std.mem.doNotOptimizeAway(&rounds);
    var round: i64 = 0;
    while (round < rounds) : (round += 1) {
        for (0..1000) |index| {
            particles.items[index] = advance(particles.items[index]);
        }
    }
    var total: i64 = 0;
    for (particles.items) |p| {
        total += p.x + p.y;
    }
    var buffer: [64]u8 = undefined;
    var writer = std.Io.File.stdout().writer(init.io, &buffer);
    try writer.interface.print("{d}\n", .{total});
    try writer.interface.flush();
}
