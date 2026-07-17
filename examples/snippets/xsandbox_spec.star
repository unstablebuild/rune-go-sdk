# xsandbox contract for the snippets example.
#
# This spec showcases how xsandbox can be used to e2e test an extension.
expect_metadata(
    id = "snippets",
    permissions = [
        "permcmd",
        "permed",
        "permfs",
        "permnoti",
        "permopen",
        "permstore",
        "permwm",
    ],
)

# Startup subscribes to selection/flush/close events and registers the command.
expect_rpc("text.Editor/SubscribeEvent")
expect_rpc("text.Editor/SubscribeCommand")
snippets = expect_command("snippets")

# A selection is remembered and `copy` opens a scratch buffer for the named
# snippet, switches the command's window to it, and tells the user how to save.
publish_event("selection", content = "print('copied')\n")
invoke_command(snippets, args = ["copy", "greeting"])
expect_rpc("workspace.Scheme/URI", where = {"path": contains("greeting.snippet")})
expect_rpc(
    "browser.ResourceOpener/Open",
    where = {"resource": contains("greeting.snippet")},
)
expect_rpc("browser.WindowManager/SetContent")
expect_rpc(
    "browser.Notifications/Notify",
    where = {"msg": "editing snippet 'greeting' (save to persist)"},
)

# Deleting a snippet removes it from persistent storage and confirms success.
invoke_command(snippets, args = ["delete", "greeting"])
expect_rpc("proto.DocumentStore/Delete", where = {"id": "greeting"})
expect_rpc(
    "browser.Notifications/Notify",
    where = {"msg": "deleted snippet 'greeting'"},
)

wait_idle("200ms")
