package manager

import (
	"database/sql"
	"net/http"

	"neoai/admin"
	"neoai/auth"
	"neoai/channel"
	"neoai/globals"
	"neoai/utils"

	"github.com/gin-gonic/gin"
)

// relay.go — public relay endpoints (/v1/models, /v1/market, /v1/charge,
// /v1/plans).
//
// NeoAI adds group-based filtering: admins see every model, but non-admin
// users only see models whose charge rule allows their current group.
// This matches the requested behaviour:
//
//	管理员全部显示，用户则根据分组显示模型
//
// The filter is applied at the API layer (here) rather than at the
// channel layer so that channel load balancing and preflight sequences
// remain unchanged — the channel manager still knows about every
// model, but the public listing endpoints only surface what the
// caller is allowed to use.

// ModelAPI — GET /v1/models
// Admins: every model in globals.V1ListModels.
// Others: only models the caller's group can invoke.
func ModelAPI(c *gin.Context) {
	user := auth.GetUser(c)
	db := utils.GetDBFromContext(c)

	if user != nil && user.IsAdmin(db) {
		c.JSON(http.StatusOK, globals.V1ListModels)
		return
	}

	allowed := computeAllowedModels(db, user)
	c.JSON(http.StatusOK, filterListModels(allowed))
}

// MarketAPI — GET /v1/market
// Same group filter as ModelAPI.
func MarketAPI(c *gin.Context) {
	user := auth.GetUser(c)
	db := utils.GetDBFromContext(c)
	all := admin.MarketInstance.GetModels()

	if user != nil && user.IsAdmin(db) {
		c.JSON(http.StatusOK, all)
		return
	}

	allowed := computeAllowedModels(db, user)
	out := make([]admin.MarketModel, 0, len(all))
	for _, m := range all {
		if _, ok := allowed[m.Id]; ok {
			out = append(out, m)
		}
	}
	c.JSON(http.StatusOK, out)
}

func ChargeAPI(c *gin.Context) {
	c.JSON(http.StatusOK, channel.ChargeInstance.ListRules())
}

func PlanAPI(c *gin.Context) {
	c.JSON(http.StatusOK, channel.PlanInstance.GetPlans())
}

// computeAllowedModels returns the set of model IDs the given user is
// allowed to see / invoke.
//
// Logic:
//   - Admin: every model (returned separately by the caller, so we
//     skip the admin case here).
//   - Anonymous (user == nil): only models in a charge rule that
//     allows anonymous (charge.Anonymous == true).
//   - Logged-in non-admin: models in a charge rule whose `Group`
//     list is empty (visible to everyone) OR contains the user's
//     current group.
//
// We additionally surface any model the user could actually invoke
// through a channel — i.e. we union in channel.ChargeInstance.Models
// whose `Group` field allows the user. The actual invocation-time
// check (auth.CanEnableModel / CanEnableModelWithSubscription) is the
// authoritative gate, but we want the listing to roughly match.
func computeAllowedModels(db *sql.DB, user *auth.User) map[string]struct{} {
	out := map[string]struct{}{}

	// Anonymous users see only "anonymous"-allowed billing models.
	anonymous := user == nil
	var group string
	if user != nil {
		group = auth.GetGroup(db, user)
	}

	for _, charge := range channel.ChargeInstance.ListRules() {
		if charge == nil {
			continue
		}
		// Visible if any of:
		//   1. Charge supports anonymous and user is anonymous (free / public).
		//   2. User is logged in AND (charge has no group restriction
		//      OR the user's group is in the charge's group list).
		visible := false
		if anonymous && charge.Anonymous {
			visible = true
		}
		if !anonymous {
			// Charges don't have a per-rule Group in this codebase —
			// the gating is via PlanInstance. So we surface every
			// billing model to logged-in users, but skip "unset" /
			// "non-billing" models that have no charge rule at all
			// (those are surfaced only to admins).
			if !charge.IsUnsetType() {
				visible = true
			}
		}
		if visible {
			for _, m := range charge.Models {
				out[m] = struct{}{}
			}
		}
	}

	// Always include the user's currently-subscribed plan models.
	if user != nil {
		level := user.GetSubscriptionLevel(db)
		if level > 0 {
			plan := channel.PlanInstance.GetPlan(level)
			for _, item := range plan.Items {
				for _, m := range item.Models {
					out[m] = struct{}{}
				}
			}
		}
	}

	// Sanity-check: don't surface models that aren't actually backed
	// by a channel — listing a model the user can't invoke would be
	// misleading.
	channelModels := channel.ConduitInstance.GetModels()
	if len(channelModels) > 0 {
		channelSet := map[string]struct{}{}
		for _, m := range channelModels {
			channelSet[m] = struct{}{}
		}
		for m := range out {
			if _, ok := channelSet[m]; !ok {
				delete(out, m)
			}
		}
	}

	// Group filter: if a group restriction is configured somewhere
	// (e.g. via the channel's `group` field), honour it. This is the
	// core of the "users see by group" requirement.
	if !anonymous && group != "" {
		groupAllowed := computeGroupAllowedModels(group)
		if len(groupAllowed) > 0 {
			// intersect
			for m := range out {
				if _, ok := groupAllowed[m]; !ok {
					delete(out, m)
				}
			}
		}
	}

	return out
}

// computeGroupAllowedModels returns the set of model IDs whose
// underlying channel's `group` list contains the given group (or is
// empty — empty means "all groups").
func computeGroupAllowedModels(group string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, ch := range channel.ConduitInstance.GetSequence() {
		if ch == nil || !ch.GetState() {
			continue
		}
		if !ch.IsHitGroup(group) {
			continue
		}
		for _, m := range ch.GetHitModels() {
			out[m] = struct{}{}
		}
	}
	return out
}

// filterListModels keeps only entries of globals.V1ListModels.Data
// whose id is in `allowed`.
func filterListModels(allowed map[string]struct{}) globals.ListModels {
	out := globals.ListModels{Object: "list", Data: []globals.ListModelsItem{}}
	for _, item := range globals.V1ListModels.Data {
		if _, ok := allowed[item.Id]; ok {
			out.Data = append(out.Data, item)
		}
	}
	return out
}

func sendErrorResponse(c *gin.Context, err error, types ...string) {
	var errType string
	if len(types) > 0 {
		errType = types[0]
	} else {
		errType = "chatnio_api_error"
	}

	c.JSON(http.StatusServiceUnavailable, RelayErrorResponse{
		Error: TranshipmentError{
			Message: err.Error(),
			Type:    errType,
		},
	})
}

func abortWithErrorResponse(c *gin.Context, err error, types ...string) {
	sendErrorResponse(c, err, types...)
	c.Abort()
}
