package main

import (
	"net/http"
	"net/url"
	"encoding/json"
	"strings"
)

func (s *server) mountAsana(mux *http.ServeMux) {
	mux.HandleFunc("GET /asana/projects", s.handleAsanaProjects)
	//tasks for a specific project with GET /asana/projects/{gid}/tasks
	mux.HandleFunc("GET /asana/projects/{gid}/tasks", s.handleAsanaProjectTasks)
}

//GET /asana/projects
func (s *server) handleAsanaProjects(w http.ResponseWriter, r *http.Request) {
	//first, get the user's workspaces
	var me struct{ Data struct {
		Workspaces []struct{ Gid, Name string } `json:"workspaces"`
	} `json:"data"` }
	if err := s.asanaGET(r, "/users/me", url.Values{"opt_fields":{"workspaces.name"}}, &me); err != nil {
		http.Error(w, err.Error(), 502); return
	}
	ws := ""
	if len(me.Data.Workspaces) > 0 {
		//instead of picking the first workspace, replace with an env var
		ws = me.Data.Workspaces[0].Gid 	}

	//projects in worksapce so not quest board yet?
	var resp struct{ Data []struct{
		Gid  string `json:"gid"`
		Name string `json:"name"`
	} `json:"data"` }
	q := url.Values{
		"workspace":  {ws},
		"archived":   {"false"},
		"opt_fields": {"name"},
		"limit":      {"100"},
	}
	if err := s.asanaGET(r, "/projects", q, &resp); err != nil {
		http.Error(w, err.Error(), 502); return
	}
	writeJSON(w, resp.Data)
}

//GET /asana/projects/{gid}/tasks
func (s *server) handleAsanaProjectTasks(w http.ResponseWriter, r *http.Request) {
	gid := r.PathValue("gid")
	if gid == "" { http.Error(w, "missing project gid", 400); return }

	//ask asana for rich fields: completion, assignee, creator, section, due dates, and custom fields.
	fields := []string{
		"name","completed","completed_at","created_at","due_on","due_at",
		"assignee.gid","assignee.name",
		"created_by.gid","created_by.name",
		"permalink_url",
		//section/column lives in memberships
		"memberships.section.name","memberships.project.name",
		//custom fields like priority/bounty/
		"custom_fields.name","custom_fields.type","custom_fields.display_value",
		"custom_fields.enum_value.name","custom_fields.enum_value.color",
		"custom_fields.number_value","custom_fields.text_value",
	}

	//paginate until done (Asana uses offset)
	type task struct {
		Gid   string `json:"gid"`
		Name  string `json:"name"`
		//we will just stream the raw task across; or map to your own struct later
		//to keep this concise, let’s just return Asana's fields as-is
	}
	var all []any
	offset := ""
	for {
		q := url.Values{
			"limit":      {"50"},
			"opt_fields": {strings.Join(fields, ",")},
		}
		if offset != "" { q.Set("offset", offset) }

		var page struct{
			Data []map[string]any `json:"data"`
			NextPage *struct{ Offset string `json:"offset"` } `json:"next_page"`
		}
		if err := s.asanaGET(r, "/projects/"+gid+"/tasks", q, &page); err != nil {
			http.Error(w, err.Error(), 502); return
		}
		for _, t := range page.Data { all = append(all, t) }
		if page.NextPage == nil || page.NextPage.Offset == "" { break }
		offset = page.NextPage.Offset
	}
	writeJSON(w, all)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func extractCustom(cfList []map[string]any, want string) (string, bool) {
	for _, cf := range cfList {
		if n, _ := cf["name"].(string); n == want {
			if v, _ := cf["display_value"].(string); v != "" { return v, true }
			//fallback for typed values if needed
		}
	}
	return "", false
}

