-- plugins/task-logger.lua
-- Logs task creations to a simple text file
-- Demonstrates external interactions and hooks

local plugin = {
    id = "task-logger",
    name = "Task Logger",
    description = "Logs new tasks to a history file",
    author = "Kairo",
    version = "1.0.0",

    config = {
        log_file = "task_history.txt"
    }
}

-- Note: Since GopherLua by default doesn't have 'io' or 'os' file access 
-- in a strictly sandboxed environment, we might need to rely on kairo 
-- if it exposed a logging/file API. 
-- However, Kairo currently doesn't expose a file API to Lua for safety.
-- For this demonstration, we'll use kairo.notify to log to the UI.

kairo.on("task_create", function(event)
    local task = event.task
    if not task then return end

    local log_msg = string.format("[%s] Created task: %s", task.id, task.title)
    
    -- In a real scenario, we'd write to a file here
    -- print(log_msg) -- GopherLua print goes to stdout
    
    kairo.notify("Logged: " .. task.title, false)
end)

kairo.on("task_delete", function(event)
    local id = event.payload.task_id
    kairo.notify("Task deleted: " .. id, false)
end)

return plugin
