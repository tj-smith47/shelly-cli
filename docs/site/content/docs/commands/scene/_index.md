---
title: "shelly scene"
description: "shelly scene"
weight: 500
sidebar:
  collapsed: true
---

## shelly scene

Manage device scenes

### Synopsis

Manage saved device state configurations (scenes).

Scenes allow you to save and recall specific device states with a single command.
Each scene contains one or more actions that are executed when the scene is activated.

### Examples

```
  # List all scenes
  shelly scene list

  # Create a new scene
  shelly scene create movie-night

  # Show scene details
  shelly scene show movie-night

  # Activate a scene
  shelly scene activate movie-night

  # Export a scene to file
  shelly scene export movie-night scene.yaml

  # Import a scene from file
  shelly scene import scene.yaml

  # Delete a scene
  shelly scene delete movie-night
```

### Options

```
  -h, --help   help for scene
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly scene activate](shelly_scene_activate.md)	 - Activate a scene
* [shelly scene create](shelly_scene_create.md)	 - Create a new scene
* [shelly scene delete](shelly_scene_delete.md)	 - Delete a scene
* [shelly scene export](shelly_scene_export.md)	 - Export a scene to file
* [shelly scene import](shelly_scene_import.md)	 - Import a scene from file
* [shelly scene list](shelly_scene_list.md)	 - List scenes
* [shelly scene show](shelly_scene_show.md)	 - Show scene details

