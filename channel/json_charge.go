package channel

import (
        "encoding/json"
        "fmt"
        "net/http"

        "neoai/globals"
        "neoai/utils"

        "github.com/gin-gonic/gin"
)

// json_charge.go — bulk charge rule import/export via a JSON document.
//
// Existing endpoints operate on a single charge rule at a time. The
// JSON bulk endpoint matches the QuantumNous/new-api style: the admin
// pastes / uploads a JSON document describing every model's pricing,
// and the backend applies it atomically (replacing or merging the
// existing rules depending on the `mode` parameter).
//
//   POST /admin/charge/json?mode=replace
//   POST /admin/charge/json?mode=merge
//
// Body shape (matches the existing Charge struct):
//
//   [
//     { "type": "token",     "models": ["gpt-4", "gpt-4-turbo"], "input": 0.03, "output": 0.06, "anonymous": false },
//     { "type": "times",     "models": ["dall-e-3"],              "input": 0,    "output": 0.04 },
//     { "type": "nonbilling","models": ["gpt-3.5-turbo-1106"] }
//   ]

type jsonChargeMode string

const (
        jsonChargeReplace jsonChargeMode = "replace"
        jsonChargeMerge   jsonChargeMode = "merge"
)

// ApplyChargeJSONAPI accepts a JSON document of charge rules and applies
// them in bulk. The `mode` query parameter controls whether existing
// rules are wiped first (replace) or kept and only augmented (merge).
func ApplyChargeJSONAPI(c *gin.Context) {
        if !utils.GetAdminFromContext(c) {
                c.JSON(http.StatusUnauthorized, gin.H{
                        "status": false,
                        "error":  "admin required",
                })
                return
        }

        mode := jsonChargeMode(c.DefaultQuery("mode", string(jsonChargeReplace)))
        if mode != jsonChargeReplace && mode != jsonChargeMerge {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  "mode must be 'replace' or 'merge'",
                })
                return
        }

        raw, err := c.GetRawData()
        if err != nil {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  err.Error(),
                })
                return
        }

        // Accept either a top-level array (preferred) or a wrapper
        // `{ "rules": [...] }` for forwards-compatibility with new-api style.
        var seq ChargeSequence
        if err := json.Unmarshal(raw, &seq); err != nil {
                var wrapper struct {
                        Rules ChargeSequence `json:"rules"`
                }
                if err2 := json.Unmarshal(raw, &wrapper); err2 == nil && wrapper.Rules != nil {
                        seq = wrapper.Rules
                } else {
                        c.JSON(http.StatusOK, gin.H{
                                "status": false,
                                "error":  fmt.Sprintf("invalid JSON: %s", err.Error()),
                        })
                        return
                }
        }

        // Validate every entry — bad rows are rejected with a useful error.
        for i, charge := range seq {
                if charge == nil {
                        c.JSON(http.StatusOK, gin.H{
                                "status": false,
                                "error":  fmt.Sprintf("rule #%d is null", i),
                        })
                        return
                }
                if len(charge.Models) == 0 {
                        c.JSON(http.StatusOK, gin.H{
                                "status": false,
                                "error":  fmt.Sprintf("rule #%d has no models", i),
                        })
                        return
                }
                // normalise type — accept both the full internal name
                // ("token-billing") and the short forms used by the
                // frontend / new-api style JSON ("token", "times",
                // "nonbilling", "free").
                switch charge.Type {
                case globals.NonBilling, globals.TimesBilling, globals.TokenBilling:
                        // ok — already canonical
                case "":
                        charge.Type = globals.TokenBilling
                case "token", "tokens":
                        charge.Type = globals.TokenBilling
                case "times", "per-call":
                        charge.Type = globals.TimesBilling
                case "nonbilling", "free":
                        charge.Type = globals.NonBilling
                default:
                        c.JSON(http.StatusOK, gin.H{
                                "status": false,
                                "error":  fmt.Sprintf("rule #%d has invalid type %q (expected: token, times, non-billing)", i, charge.Type),
                        })
                        return
                }
                // reset id so each row gets a fresh slot in replace mode
                if mode == jsonChargeReplace {
                        charge.Id = -1
                }
        }

        if mode == jsonChargeReplace {
                // Wipe everything first by replacing the entire sequence.
                ChargeInstance.Sequence = ChargeSequence{}
        }

        for _, charge := range seq {
                ChargeInstance.AddRawRule(charge)
        }

        if err := ChargeInstance.SaveConfig(); err != nil {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "status": true,
                "data":   ChargeInstance.ListRules(),
        })
}

// ExportChargeJSONAPI returns the current charge rules as a JSON array.
// Useful for backup or for editing in an external editor before re-import.
func ExportChargeJSONAPI(c *gin.Context) {
        if !utils.GetAdminFromContext(c) {
                c.JSON(http.StatusUnauthorized, gin.H{
                        "status": false,
                        "error":  "admin required",
                })
                return
        }
        c.JSON(http.StatusOK, gin.H{
                "status": true,
                "data":   ChargeInstance.ListRules(),
        })
}
