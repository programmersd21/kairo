return {
    id = "sample",
    name = "Sample Plugin",
    description = "Demonstrates the expanded Kairo Lua API",
    author = "Kairo Team",
    version = "1.0.0",

    commands = {
        {
            id = "hello",
            title = "Lua: Hello World",
            hint = "shows a notification",
            run = function()
                kairo.notify("Greetings from the Lua Sample Plugin!", false)
            end
        },
        {
            id = "cleanup-done",
            title = "Lua: Delete Done Tasks",
            hint = "removes all completed tasks",
            run = function()
                local tasks = kairo.list_tasks({statuses = {"done"}})
                local count = 0
                for _, t in ipairs(tasks) do
                    kairo.delete_task(t.id)
                    count = count + 1
                end
                kairo.notify("Cleaned up " .. count .. " tasks", false)
            end
        }
    },

    views = {
        {
            id = "doing-now",
            title = "Active Work",
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
