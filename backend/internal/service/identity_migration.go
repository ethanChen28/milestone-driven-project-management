package service

import "sort"

type IdentityReference struct {
	ObjectType string `json:"objectType"`
	ObjectID   string `json:"objectId"`
	Field      string `json:"field"`
	Value      string `json:"value"`
	ResolvedID string `json:"resolvedId,omitempty"`
	Resolved   bool   `json:"resolved"`
}

type IdentityMigrationReport struct {
	TotalReferences      int                 `json:"totalReferences"`
	ResolvedReferences   int                 `json:"resolvedReferences"`
	UnresolvedReferences int                 `json:"unresolvedReferences"`
	References           []IdentityReference `json:"references"`
}

func (s *Store) IdentityMigrationReport(users []UserProfile) IdentityMigrationReport {
	lookup := map[string]string{}
	for _, user := range users {
		lookup[user.ID] = user.ID
		lookup[user.Username] = user.ID
		lookup[user.Email] = user.ID
	}
	report := IdentityMigrationReport{}
	s.mu.RLock()
	defer s.mu.RUnlock()
	add := func(objectType, objectID, field, value string) {
		if value == "" {
			return
		}
		ref := IdentityReference{ObjectType: objectType, ObjectID: objectID, Field: field, Value: value}
		if id, ok := lookup[value]; ok {
			ref.Resolved = true
			ref.ResolvedID = id
			report.ResolvedReferences++
		} else {
			report.UnresolvedReferences++
		}
		report.TotalReferences++
		report.References = append(report.References, ref)
	}
	for _, project := range s.projects {
		add("project", project.ID, "owner", project.Owner)
		for _, participant := range project.Participants {
			add("project", project.ID, "participants", participant)
		}
	}
	for _, milestone := range s.milestones {
		add("milestone", milestone.ID, "owner", milestone.Owner)
	}
	for _, item := range s.workItems {
		add("work_item", item.ID, "owner", item.Owner)
	}
	for _, update := range s.updates {
		add("weekly_update", update.ID, "author", update.Author)
	}
	sort.Slice(report.References, func(i, j int) bool {
		if report.References[i].ObjectType == report.References[j].ObjectType {
			return report.References[i].ObjectID < report.References[j].ObjectID
		}
		return report.References[i].ObjectType < report.References[j].ObjectType
	})
	return report
}
