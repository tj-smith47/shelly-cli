// Package automation provides script, schedule, and event automation for Shelly devices.
package automation

import (
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// BuiltInScriptTemplates returns the bundled script templates.
// These are maintained by the CLI and provide common functionality.
func BuiltInScriptTemplates() map[string]config.ScriptTemplate {
	return map[string]config.ScriptTemplate{
		"motion-light": {
			Name:        "motion-light",
			Description: "Turn on light when motion is detected, auto-off after timeout",
			Category:    "automation",
			MinGen:      2,
			BuiltIn:     true,
			Author:      "Shelly CLI",
			Version:     "1.0.0",
			Variables: []config.ScriptVariable{
				{Name: "LIGHT_ID", Description: "Light component ID (0-3)", Type: "number", Default: 0, Required: true},
				{Name: "INPUT_ID", Description: "Motion sensor input ID (0-3)", Type: "number", Default: 0, Required: true},
				{Name: "TIMEOUT_SEC", Description: "Auto-off timeout in seconds", Type: "number", Default: 300, Required: false},
			},
			Code: `// Motion-activated light control
// Turns on light when motion detected, auto-off after timeout

let CONFIG = {
  lightId: LIGHT_ID,      // Light component ID
  inputId: INPUT_ID,      // Motion sensor input ID
  timeout: TIMEOUT_SEC    // Auto-off delay in seconds
};

let timer = null;

function turnOffLight() {
  Shelly.call("Light.Set", {id: CONFIG.lightId, on: false});
  print("Motion timeout - light off");
}

Shelly.addEventHandler(function(event) {
  if (event.component === "input:" + CONFIG.inputId) {
    if (event.info.state === true) {
      // Motion detected
      if (timer !== null) {
        Timer.clear(timer);
      }
      Shelly.call("Light.Set", {id: CONFIG.lightId, on: true});
      print("Motion detected - light on");
      timer = Timer.set(CONFIG.timeout * 1000, false, turnOffLight);
    }
  }
});

print("Motion light script started");
`,
		},
		"power-monitor": {
			Name:        "power-monitor",
			Description: "Monitor power consumption and log high usage alerts",
			Category:    "monitoring",
			MinGen:      2,
			BuiltIn:     true,
			Author:      "Shelly CLI",
			Version:     "1.0.0",
			Variables: []config.ScriptVariable{
				{Name: "SWITCH_ID", Description: "Switch component ID (0-3)", Type: "number", Default: 0, Required: true},
				{Name: "THRESHOLD_W", Description: "Power threshold in watts for alert", Type: "number", Default: 1000, Required: false},
				{Name: "CHECK_INTERVAL_SEC", Description: "Check interval in seconds", Type: "number", Default: 60, Required: false},
			},
			Code: `// Power consumption monitor
// Logs alerts when power exceeds threshold

let CONFIG = {
  switchId: SWITCH_ID,
  threshold: THRESHOLD_W,
  interval: CHECK_INTERVAL_SEC
};

function checkPower() {
  Shelly.call("Switch.GetStatus", {id: CONFIG.switchId}, function(result, error) {
    if (error) {
      print("Error getting status: " + error);
      return;
    }
    let power = result.apower || 0;
    if (power > CONFIG.threshold) {
      print("HIGH POWER ALERT: " + power.toFixed(1) + "W exceeds " + CONFIG.threshold + "W threshold");
    }
  });
}

Timer.set(CONFIG.interval * 1000, true, checkPower);
print("Power monitor started - checking every " + CONFIG.interval + "s, threshold: " + CONFIG.threshold + "W");
`,
		},
		"schedule-helper": {
			Name:        "schedule-helper",
			Description: "Simple on/off scheduler with sunrise/sunset support",
			Category:    "automation",
			MinGen:      2,
			BuiltIn:     true,
			Author:      "Shelly CLI",
			Version:     "1.0.0",
			Variables: []config.ScriptVariable{
				{Name: "SWITCH_ID", Description: "Switch component ID (0-3)", Type: "number", Default: 0, Required: true},
				{Name: "ON_HOUR", Description: "Hour to turn on (0-23, or -1 for sunset)", Type: "number", Default: 18, Required: false},
				{Name: "OFF_HOUR", Description: "Hour to turn off (0-23, or -1 for sunrise)", Type: "number", Default: 23, Required: false},
			},
			Code: `// Simple scheduler
// Turn switch on/off at specified hours

let CONFIG = {
  switchId: SWITCH_ID,
  onHour: ON_HOUR,
  offHour: OFF_HOUR
};

function checkSchedule() {
  let now = new Date();
  let hour = now.getHours();

  Shelly.call("Switch.GetStatus", {id: CONFIG.switchId}, function(result, error) {
    if (error) return;

    let isOn = result.output;
    let shouldBeOn = false;

    if (CONFIG.onHour <= CONFIG.offHour) {
      shouldBeOn = hour >= CONFIG.onHour && hour < CONFIG.offHour;
    } else {
      shouldBeOn = hour >= CONFIG.onHour || hour < CONFIG.offHour;
    }

    if (shouldBeOn && !isOn) {
      Shelly.call("Switch.Set", {id: CONFIG.switchId, on: true});
      print("Schedule: turning ON at hour " + hour);
    } else if (!shouldBeOn && isOn) {
      Shelly.call("Switch.Set", {id: CONFIG.switchId, on: false});
      print("Schedule: turning OFF at hour " + hour);
    }
  });
}

Timer.set(60000, true, checkSchedule);
checkSchedule();
print("Scheduler started: ON at " + CONFIG.onHour + ":00, OFF at " + CONFIG.offHour + ":00");
`,
		},
		"toggle-sync": {
			Name:        "toggle-sync",
			Description: "Sync state between two switches (master/slave)",
			Category:    "automation",
			MinGen:      2,
			BuiltIn:     true,
			Author:      "Shelly CLI",
			Version:     "1.0.0",
			Variables: []config.ScriptVariable{
				{Name: "MASTER_ID", Description: "Master switch component ID", Type: "number", Default: 0, Required: true},
				{Name: "SLAVE_ID", Description: "Slave switch component ID", Type: "number", Default: 1, Required: true},
			},
			Code: `// Switch synchronization
// Slave switch follows master switch state

let CONFIG = {
  masterId: MASTER_ID,
  slaveId: SLAVE_ID
};

Shelly.addEventHandler(function(event) {
  if (event.component === "switch:" + CONFIG.masterId) {
    if (typeof event.info.output !== "undefined") {
      Shelly.call("Switch.Set", {id: CONFIG.slaveId, on: event.info.output});
      print("Synced slave to " + (event.info.output ? "ON" : "OFF"));
    }
  }
});

print("Toggle sync active: switch " + CONFIG.masterId + " -> switch " + CONFIG.slaveId);
`,
		},
		"energy-logger": {
			Name:        "energy-logger",
			Description: "Log hourly energy consumption to KVS for tracking",
			Category:    "monitoring",
			MinGen:      2,
			BuiltIn:     true,
			Author:      "Shelly CLI",
			Version:     "1.0.0",
			Variables: []config.ScriptVariable{
				{Name: "SWITCH_ID", Description: "Switch component ID with energy metering", Type: "number", Default: 0, Required: true},
			},
			Code: `// Energy consumption logger
// Stores hourly readings in KVS

let CONFIG = {
  switchId: SWITCH_ID
};

let lastEnergy = 0;

function logEnergy() {
  Shelly.call("Switch.GetStatus", {id: CONFIG.switchId}, function(result, error) {
    if (error) {
      print("Error: " + error);
      return;
    }

    let energy = result.aenergy ? result.aenergy.total : 0;
    let now = new Date();
    let key = "energy_" + now.getFullYear() + "_" +
              (now.getMonth() + 1).toString().padStart(2, "0") + "_" +
              now.getDate().toString().padStart(2, "0") + "_" +
              now.getHours().toString().padStart(2, "0");

    let consumption = energy - lastEnergy;
    if (lastEnergy > 0 && consumption >= 0) {
      Shelly.call("KVS.Set", {key: key, value: consumption.toFixed(2)});
      print("Logged " + consumption.toFixed(2) + "Wh for " + key);
    }
    lastEnergy = energy;
  });
}

// Initialize and start hourly logging
Shelly.call("Switch.GetStatus", {id: CONFIG.switchId}, function(result, error) {
  if (!error && result.aenergy) {
    lastEnergy = result.aenergy.total;
  }
});

Timer.set(3600000, true, logEnergy);
print("Energy logger started for switch " + CONFIG.switchId);
`,
		},
	}
}

// GetScriptTemplate returns a script template by name, checking built-in templates first.
func GetScriptTemplate(name string) (config.ScriptTemplate, bool) {
	// Check built-in templates first
	builtIn := BuiltInScriptTemplates()
	if tpl, ok := builtIn[name]; ok {
		return tpl, true
	}

	// Check user-defined templates
	return config.GetScriptTemplate(name)
}

// ListAllScriptTemplates returns all script templates (built-in + user-defined).
func ListAllScriptTemplates() map[string]config.ScriptTemplate {
	result := make(map[string]config.ScriptTemplate)

	// Add built-in templates
	for name, tpl := range BuiltInScriptTemplates() {
		result[name] = tpl
	}

	// Add user-defined templates (can override built-in)
	for name, tpl := range config.ListScriptTemplates() {
		result[name] = tpl
	}

	return result
}

// SubstituteVariables replaces variable placeholders in template code.
func SubstituteVariables(code string, values map[string]any) string {
	for name, value := range values {
		var replacement string
		switch v := value.(type) {
		case string:
			replacement = fmt.Sprintf("%q", v)
		case nil:
			replacement = "null"
		default:
			replacement = fmt.Sprintf("%v", v)
		}
		code = strings.ReplaceAll(code, name, replacement)
	}
	return code
}
