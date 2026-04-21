-- plugins/auto-cleanup.lua
-- Automatically removes completed tasks older than N days
-- Demonstrates filtering, batch operations, and configuration

local plugin = {
    id = "auto-cleanup",
    name = "Auto Cleanup",
    description = "Periodically removes completed tasks older than a configurable threshold",
    author = "Kairo",
    version = "1.0.1",
    
    config = {
        days_threshold = 30,  -- Delete completed tasks older than 30 days
        auto_run = true,      -- Set to true to run on app start
    }
}

-- Manual cleanup command
local function cleanup_old_tasks()
    -- Get all completed tasks
    local tasks, err = kairo.list_tasks({
        statuses = {"done"},
        sort = "updated"
    })
    
    if err then
        kairo.notify("Failed to list tasks: " .. err, true)
        return
    end

    local now = os.time()
    local threshold = plugin.config.days_threshold * 24 * 3600
    local deleted_count = 0

    for _, task in ipairs(tasks) do
        -- Parse ISO8601 timestamp (very simplified)
        -- Kairo returns "2006-01-02T15:04:05Z"
        local year, month, day, hour, min, sec = 
            string.match(task.updated_at, "(%d+)-(%d+)-(%d+)T(%d+):(%d+):(%d+)")
        
        if year then
            local task_time = os.time({
                year = tonumber(year),
                month = tonumber(month),
                day = tonumber(day),
                hour = tonumber(hour),
                min = tonumber(min),
                sec = tonumber(sec)
            })
            
            if now - task_time > threshold then
                local ok, err = kairo.delete_task(task.id)
                if ok then
                    deleted_count = deleted_count + 1
                end
            end
        end
    end

    if deleted_count > 0 then
        local msg = string.format("Cleanup complete: deleted %d old tasks", deleted_count)
        kairo.notify(msg, false)
    end
end

-- Hook into app startup if enabled
if plugin.config.auto_run then
    kairo.on("app_start", function(event)
        cleanup_old_tasks()
    end)
end

-- Expose commands and views
plugin.commands = {
    {
        id = "cleanup",
        title = "Cleanup: Remove Old Done Tasks",
        hint = "sweep",
        run = cleanup_old_tasks
    }
}

plugin.views = {
    {
        id = "old-tasks",
        title = "Older Done Tasks",
        filter = {
            statuses = {"done"},
            sort = "updated"
        }
    }
}

return plugin
