-- plugins/sample.lua
-- This is a sample plugin for Kairo demonstrating the expanded Lua API.
-- Plugins allow you to extend Kairo with custom commands, views, and automation.

local plugin = {
    -- Unique identifier for the plugin
    id = "sample",
    -- Display name in the plugin menu
    name = "Sample Plugin",
    -- Brief description of what the plugin does
    description = "Demonstrates the expanded Kairo Lua API",
    -- Metadata for the user
    author = "Kairo",
    version = "1.0.0",

    -- Commands are actions that appear in the Command Palette (ctrl+p)
    commands = {
        {
            -- Unique ID for this command
            id = "hello",
            -- Text shown in the palette
            title = "Lua: Hello World",
            -- Small hint text shown next to the title
            hint = "shows a notification",
            -- Function called when the command is selected
            run = function()
                -- kairo.notify(message, is_error) sends a notification to the TUI status bar
                kairo.notify("Greetings from the Lua Sample Plugin!", false)
            end
        },
        {
            id = "cleanup-done",
            title = "Lua: Delete Done Tasks",
            hint = "removes all completed tasks",
            run = function()
                -- kairo.list_tasks(filter) retrieves tasks matching the criteria
                -- filter keys: statuses (table), tag (string), priority (number), sort (string)
                local tasks = kairo.list_tasks({statuses = {"done"}})
                local count = 0
                for _, t in ipairs(tasks) do
                    -- kairo.delete_task(id) removes a task by ID
                    kairo.delete_task(t.id)
                    count = count + 1
                end
                kairo.notify("Cleaned up " .. count .. " tasks", false)
            end
        }
    },

    -- Views are custom tabs that appear at the top of the main task list
    views = {
        {
            -- Unique ID for this view
            id = "doing-now",
            -- Title shown on the tab
            title = "Active Work",
            -- Filter criteria for tasks shown in this view
            filter = {
                statuses = {"doing"},
                sort = "updated"
            }
        },
        {
            id = "high-pri-work",
            title = "Critical",
            filter = {
                statuses = {"todo", "doing"},
                min_priority = 0,
                sort = "deadline"
            }
        }
    }
}

-- Event hooks allow you to react to app-wide events automatically
-- Supported events: task_create, task_update, task_delete, app_start, app_stop

kairo.on("task_create", function(event)
    -- event.task contains the newly created task table
    local task = event.task
    if task and task.title:match(" urgent") then
        kairo.notify("Urgent task detected: " .. task.title, true)
    end
end)

-- Plugins must return the table containing metadata, commands, and views
return plugin
