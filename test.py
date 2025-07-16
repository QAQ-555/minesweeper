import random
from collections import deque

# 地图尺寸
MAP_WIDTH = 20
MAP_HEIGHT = 20

# 地图初始化，0表示空地，1表示山体，2表示河流
map_grid = [[0 for _ in range(MAP_WIDTH)] for _ in range(MAP_HEIGHT)]

# 连通性检查：使用广度优先搜索 (BFS) 来检查地图是否连通
def is_connected(map_grid):
    visited = [[False for _ in range(MAP_WIDTH)] for _ in range(MAP_HEIGHT)]
    
    # 寻找一个空地作为起始点
    start = None
    for y in range(MAP_HEIGHT):
        for x in range(MAP_WIDTH):
            if map_grid[y][x] == 0:  # 找到一个空地
                start = (x, y)
                break
        if start:
            break

    if not start:
        return False  # 没有空地，地图本身就被完全占据

    # 广度优先搜索（BFS）
    queue = deque([start])
    visited[start[1]][start[0]] = True
    directions = [(-1, 0), (1, 0), (0, -1), (0, 1)]  # 上、下、左、右四个方向

    while queue:
        x, y = queue.popleft()

        for dx, dy in directions:
            nx, ny = x + dx, y + dy
            if 0 <= nx < MAP_WIDTH and 0 <= ny < MAP_HEIGHT and not visited[ny][nx] and map_grid[ny][nx] == 0:
                visited[ny][nx] = True
                queue.append((nx, ny))

    # 如果所有空地都被访问过，说明是连通的
    for y in range(MAP_HEIGHT):
        for x in range(MAP_WIDTH):
            if map_grid[y][x] == 0 and not visited[y][x]:
                return False  # 如果有空地没有被访问到，说明地图不连通

    return True

# 随机生成障碍物，保证地图大致连通
def generate_obstacles(map_grid, obstacle_density=0.3):
    # 随机生成障碍物
    for y in range(MAP_HEIGHT):
        for x in range(MAP_WIDTH):
            if random.random() < obstacle_density:
                # 随机选择山体（1）或河流（2）
                map_grid[y][x] = random.choice([1, 2])

# 生成并检查地图连通性
def generate_map():
    while True:
        # 生成地图并初始化为空地
        map_grid = [[0 for _ in range(MAP_WIDTH)] for _ in range(MAP_HEIGHT)]
        
        # 随机生成障碍物
        generate_obstacles(map_grid, obstacle_density=0.3)
        
        # 检查连通性
        if is_connected(map_grid):
            break  # 如果地图连通，结束生成

    return map_grid

# 打印生成的地图
def print_map(map_grid):
    for row in map_grid:
        print(" ".join(str(cell) for cell in row))

# 生成并打印地图
map_grid = generate_map()
print_map(map_grid)
