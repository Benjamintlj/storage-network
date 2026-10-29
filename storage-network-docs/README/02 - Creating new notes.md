---
cssclasses:
  - teal
  - plain
parent: "[[Readme]]"
---

## Structure of a note

Practically all notes should be made via templates, found at `99 - Meta/00 - Templates`.
All notes should contain a properties table at the top, with a `cssclasses` field.

I have created a custom css class called `plain`, which contains colour sets red, blue, green, purple, orange, teal & yellow. 

This properties table should be used for all linking between tables, at the minimum all templates will contain a `parent` field, which is used to link to parent note of that note.

Okay so why bother linking to the `parent` note, and the answer is... because it looks cool on the open graph view. 

It would be good practice to fill in these properties each time you create a new note, this will ensure that your note do remain organised.

## Creating a Quick Note

1. `cmd/ctrl` + `p`
2. Type: *"Daily notes: Open today's note"*

You can update the template at [[Quick Note]].

## Creating a Ticket or Epic

1. Go to the folder that you want to add to
2. Type add note
3. Call it what you want
4. `cmd/ctrl` + `p`, *"Insert template"*
5. Select your template
6. `cmd/ctrl` + `p`, *"Replace templates in active file"*
*Note: You can create shortcuts for these commands in settings*

## Creating a Service

1. In the file explorer right click the folder you want to create a new note in
2. Select *"create new note from template"*
3. Select service