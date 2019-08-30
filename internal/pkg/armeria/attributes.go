package armeria

const (
	AttributeChannels    string = "channels"
	AttributeColor       string = "color"
	AttributeDescription string = "description"
	AttributePermissions string = "permissions"
	AttributePicture     string = "picture"
	AttributeRarity      string = "rarity"
	AttributeScript      string = "script"
	AttributeTitle       string = "title"
	AttributeType        string = "type"
	AttributeNorth       string = "north"
	AttributeEast        string = "east"
	AttributeSouth       string = "south"
	AttributeWest        string = "west"
	AttributeUp          string = "up"
	AttributeDown        string = "down"
)

const (
	TempAttributeEditorOpen string = "editorOpen"
	TempAttributeGhost      string = "ghost"
	TempAttributeReplyTo    string = "replyTo"
)

// ValidAreaAttributes returns an array of valid attributes that can be permanently set.
func ValidAreaAttributes() []string {
	return []string{}
}

// ValidCharacterAttributes returns an array of valid attributes that can be permanently set.
func ValidCharacterAttributes() []string {
	return []string{
		AttributePicture,
		AttributeTitle,
		AttributePermissions,
		AttributeChannels,
	}
}

// ValidItemAttributes returns an array of valid attributes that can be permanently set.
func ValidItemAttributes() []string {
	return []string{
		AttributePicture,
		AttributeRarity,
	}
}

// ValidItemInstanceAttributes returns an array of attributes that can be overriden from the parent.
func ValidItemInstanceAttributes() []string {
	return []string{
		AttributeRarity,
	}
}

// ValidMobInstanceAttributes returns an array of attributes that can be overriden from the parent.
func ValidMobInstanceAttributes() []string {
	return []string{}
}

// ValidMobAttributes returns an array of valid attributes that can be permanently set.
func ValidMobAttributes() []string {
	return []string{
		AttributePicture,
		AttributeScript,
	}
}

// ValidRoomAttributes returns an array of valid attributes that can be permanently set.
func ValidRoomAttributes() []string {
	return []string{
		AttributeTitle,
		AttributeDescription,
		AttributeColor,
		AttributeType,
		AttributeNorth,
		AttributeEast,
		AttributeSouth,
		AttributeWest,
		AttributeUp,
		AttributeDown,
	}
}

// AreaAttributeDefault returns the default value for a particular attribute.
func AreaAttributeDefault(name string) string {
	switch name {

	}

	return ""
}

// CharacterAttributeDefault returns the default value for a particular attribute.
func CharacterAttributeDefault(name string) string {
	switch name {

	}

	return ""
}

// ItemAttributeDefault returns the default value for a particular attribute.
func ItemAttributeDefault(name string) string {
	switch name {
	case AttributeRarity:
		return "0"
	}

	return ""
}

// MobAttributeDefault returns the default value for a particular attribute.
func MobAttributeDefault(name string) string {
	switch name {

	}

	return ""
}

// RoomAttributeDefault returns the default value for a particular attribute.
func RoomAttributeDefault(name string) string {
	switch name {
	case AttributeTitle:
		return "Empty Room"
	case AttributeDescription:
		return "You are in a newly created empty room. Make it a good one!"
	case AttributeColor:
		return "190,190,190"
	}

	return ""
}
