package main

import (
	"errors"
	"fmt"
)

func NewInstanceGroups() *InstanceGroups {
	return &InstanceGroups{groups: make(map[string]*InstanceGroup)}
}

type InstanceGroups struct {
	groups map[string]*InstanceGroup
}

func (igs *InstanceGroups) GetGroups() []string {
	result := make([]string, len(igs.groups))
	var i int = 0
	for group := range igs.groups {
		result[i] = group
		i++
	}
	return result
}

func (igs *InstanceGroups) GetMembers(groupName string) []string {
	grp, ok := igs.groups[groupName]
	if !ok {
		return []string{}
	}
	return grp.GetMembers()
}

func (igs *InstanceGroups) Delete(groupName string) {
	_, ok := igs.groups[groupName]
	if !ok {
		return
	}
	delete(igs.groups, groupName)
}

func (igs *InstanceGroups) getGroup(groupName string) (*InstanceGroup, error) {
	grp, ok := igs.groups[groupName]
	if !ok {
		return nil, errors.New(fmt.Sprintf("group %v not found", groupName))
	}
	return grp, nil
}

func (igs *InstanceGroups) HasMember(groupName string, instanceName string) bool {
	grp, err := igs.getGroup(groupName)
	if err != nil {
		return false
	}
	return grp.HasMember(instanceName)
}

func (igs *InstanceGroups) AddInstance(groupName string, instanceName string) {
	grp, err := igs.getGroup(groupName)
	if err != nil {
		grp = NewInstanceGroup(groupName)
		igs.groups[groupName] = grp
	}
	if grp.HasMember(instanceName) {
		return
	}
	grp.AddInstance(instanceName)
}

func (igs *InstanceGroups) RemoveInstance(groupName string, instanceName string) {
	grp, err := igs.getGroup(groupName)
	if err != nil {
		return
	}
	grp.RemoveInstance(instanceName)
}

type InstanceGroup struct {
	name    string
	members []string
}

func NewInstanceGroup(name string) *InstanceGroup {
	return &InstanceGroup{name: name, members: []string{}}
}

func (ig *InstanceGroup) GetMembers() []string {
	return ig.members
}

func (ig *InstanceGroup) hasMember(instanceName string) int {
	for key, m := range ig.members {
		if m == instanceName {
			return key
		}
	}
	return -1
}

func (ig *InstanceGroup) HasMember(instanceName string) bool {
	return ig.hasMember(instanceName) >= 0
}

func (ig *InstanceGroup) AddInstance(instanceName string) {
	if ig.hasMember(instanceName) >= 0 {
		return
	}
	ig.members = append(ig.members, instanceName)
}

func (ig *InstanceGroup) RemoveInstance(instanceName string) {
	key := ig.hasMember(instanceName)
	if key < 0 {
		return
	}
	ig.members = append(ig.members[:key], ig.members[key+1:]...)
}
