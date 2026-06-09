# Tasks

The fundamental unit of organization in Kairo is the **Task**. Every task is designed to capture the essence of what needs to be done with minimal friction.

<img src={require('../../assets/tasks_list.png').default} alt="Managing Tasks" />

## Task Fields

A task in Kairo contains the following fields:

| Field | Description |
|---|---|
| **Title** | A short, descriptive name for the task. |
| **Description** | Detailed notes, supporting Markdown. |
| **Status** | `todo`, `doing`, or `done`. |
| **Priority** | `0` (Critical), `1` (High), `2` (Medium), `3` (Low). |
| **Tags** | Comma-separated labels for categorization. |
| **Due Date** | An optional deadline (supports natural language). |
| **Project** | The project this task belongs to. |
| **Parent** | For nested tasks, the ID of the parent task. |
| **Result** | A completion note or countermeasure documenting the outcome. |

## Creating & Editing Tasks

Press `n` from the main list view to open the task editor for a new task, or `e` to edit an existing one.

- Use `Tab` / `Shift+Tab` to navigate between fields.
- **Pro Tip:** Press `Enter` on the **Project** field to select from existing projects.
- Press `ctrl+s` to save.
- Press `Esc` to cancel.

### Natural Language Deadlines

Kairo features a powerful NLP engine for deadlines. You can type:
- `tomorrow 10am`
- `next friday`
- `in 2 days`
- `end of month`
- `mon 3pm`

## Status Management & Completion Notes

You can quickly change task status from the list view:
- Press `z` to toggle completion.
- **Completion Notes (Countermeasures):** When you mark a task as **Done**, Kairo will prompt you for a **Result** note. This is a great place to document what was achieved, any obstacles encountered, or a "countermeasure" to prevent future issues. 
- You can edit the result note later by pressing `z` on a task that is already completed.
- Tasks in `doing` state are highlighted to indicate active focus.

## Bulk Actions

Select multiple tasks using `Space` and perform actions on all of them at once (e.g., complete, delete).
