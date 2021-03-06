package armeria

import (
	"armeria/internal/pkg/misc"
	"armeria/internal/pkg/sfx"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Force verify that Character implements ContainerObject.
var _ ContainerObject = (*Character)(nil)

// A Character is the player's logged in character.
type Character struct {
	sync.RWMutex
	UUID                 string            `json:"uuid"`
	UnsafeName           string            `json:"name"`
	UnsafePassword       string            `json:"password"`
	UnsafeAttributes     map[string]string `json:"attributes"`
	UnsafeSettings       map[string]string `json:"settings"`
	UnsafeInventory      *ObjectContainer  `json:"inventory"`
	UnsafeEquipment      *ObjectContainer  `json:"equipment"`
	UnsafeTempAttributes map[string]string `json:"-"`
	UnsafeLastSeen       time.Time         `json:"lastSeen"`
	UnsafeMobConvo       *Conversation     `json:"-"`
	player               *Player
}

// PronounType is used to determine the correct pronoun (he/she etc.)
type PronounType int

// Character constants
const (
	ColorRoomTitle int = iota
	ColorSay
	ColorMovement
	ColorMovementAlt
	ColorError
	ColorRoomDirs
	ColorWhisper
	ColorSuccess
	ColorCmdHelp
	ColorChannelGeneral
	ColorChannelCore
	ColorChannelBuilders
	ColorMoney

	PronounSubjective PronounType = iota
	PronounPossessiveAdjective
	PronounPossessiveAbsolute
	PronounObjective
)

// Init is called when the Character is created or loaded from disk.
func (c *Character) Init() {
	// Initialize the inventory, if not defined.
	if c.UnsafeInventory == nil {
		c.UnsafeInventory = NewObjectContainer(35)
	}
	// Initialize the equipment, if not defined.
	if c.UnsafeEquipment == nil {
		c.UnsafeEquipment = NewObjectContainer(0)
	}
	// Attach parents to the child containers.
	c.UnsafeInventory.AttachParent(c, ContainerParentTypeCharacter)
	c.UnsafeEquipment.AttachParent(c, ContainerParentTypeCharacter)
	// Sync the containers.
	c.UnsafeInventory.Sync()
	c.UnsafeEquipment.Sync()
	// Register the Character with global registry.
	Armeria.registry.Register(c, c.ID(), RegistryTypeCharacter)
}

// ID returns the uuid of the Character.
func (c *Character) ID() string {
	return c.UUID
}

// Type returns the object type, since Character implements the ContainerObject interface.
func (c *Character) Type() ContainerObjectType {
	return ContainerObjectTypeCharacter
}

// Name returns the raw Character name.
func (c *Character) Name() string {
	c.RLock()
	defer c.RUnlock()
	return c.UnsafeName
}

// FormattedName returns the formatted Character name.
func (c *Character) FormattedName() string {
	c.RLock()
	defer c.RUnlock()
	return TextStyle(c.UnsafeName, WithBold())
}

// FormattedNameWithTitle returns the formatted Character name including the unsafeCharacter's title (if set).
func (c *Character) FormattedNameWithTitle() string {
	c.RLock()
	defer c.RUnlock()

	title := c.UnsafeAttributes["title"]
	if title != "" {
		return fmt.Sprintf("%s &lt;%s&gt;", TextStyle(c.UnsafeName, WithBold()), title)
	}

	return TextStyle(c.UnsafeName, WithBold())
}

// CheckPassword returns a bool indicating whether the password is correct or not.
func (c *Character) CheckPassword(pw string) bool {
	c.RLock()
	defer c.RUnlock()

	byteHash := []byte(c.UnsafePassword)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(pw))
	if err != nil {
		return false
	}

	return true
}

// SetPassword hashes and sets a new password for the Character.
func (c *Character) SetPassword(pw string) {
	c.Lock()
	defer c.Unlock()

	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		Armeria.log.Fatal("error generating password hash",
			zap.Error(err),
		)
	}

	c.UnsafePassword = string(hash)
}

// PasswordHash returns the Character's already-encrypted password as an md5 hash.
func (c *Character) PasswordHash() string {
	c.RLock()
	defer c.RUnlock()

	b := []byte(c.UnsafePassword)
	return fmt.Sprintf("%x", md5.Sum(b))
}

// Inventory returns the Character's inventory.
func (c *Character) Inventory() *ObjectContainer {
	c.RLock()
	defer c.RUnlock()

	return c.UnsafeInventory
}

// Equipment returns the Character's equipment.
func (c *Character) Equipment() *ObjectContainer {
	c.RLock()
	defer c.RUnlock()

	return c.UnsafeEquipment
}

// Player returns the parent that is playing the Character.
func (c *Character) Player() *Player {
	c.RLock()
	defer c.RUnlock()

	return c.player
}

// SetPlayer sets the parent that is playing the Character.
func (c *Character) SetPlayer(p *Player) {
	c.Lock()
	defer c.Unlock()

	c.player = p
}

// MobConvo returns the active mob conversation for the Character.
func (c *Character) MobConvo() *Conversation {
	c.RLock()
	defer c.RUnlock()

	return c.UnsafeMobConvo
}

// SetMobConvo sets the active mob conversation with the Character.
func (c *Character) SetMobConvo(convo *Conversation) {
	c.Lock()
	defer c.Unlock()

	c.UnsafeMobConvo = convo
}

// Room returns the Character's Room based on the object container it is within.
func (c *Character) Room() *Room {
	oc := Armeria.registry.GetObjectContainer(c.ID())
	if oc == nil {
		return nil
	}
	return oc.ParentRoom()
}

// UserColor will return the corresponding color according to the Character's color settings.
func (c *Character) UserColor(color int) string {
	switch color {
	case ColorRoomTitle:
		return "#6e94ff"
	case ColorSay:
		return "#ffeb3b"
	case ColorMovement:
		return "#00bcd4"
	case ColorMovementAlt:
		return "#00ffc6"
	case ColorError:
		return "#e91e63"
	case ColorRoomDirs:
		return "#4c9af3"
	case ColorWhisper:
		return "#b730f7"
	case ColorSuccess:
		return "#8ee22b"
	case ColorCmdHelp:
		return "#e9761e"
	case ColorChannelGeneral:
		return "#009688"
	case ColorChannelCore:
		return "#ff5722"
	case ColorChannelBuilders:
		return "#007cff"
	case ColorMoney:
		return "#fec205"
	default:
		return ""
	}
}

// Colorize will color text according to the Character's color settings.
// TODO Deprecate Character.Colorize() in favor of using TextStyles() with WithUserColor().
func (c *Character) Colorize(text string, color int) string {
	return fmt.Sprintf("<span style='color:%s'>%s</span>", c.UserColor(color), text)
}

// LastSeen returns the Time the Character last successfully logged into the game.
func (c *Character) LastSeen() time.Time {
	c.RLock()
	defer c.RUnlock()
	return c.UnsafeLastSeen
}

// SetLastSeen sets the time the Character last logged into the game.
func (c *Character) SetLastSeen(seen time.Time) {
	c.Lock()
	defer c.Unlock()
	c.UnsafeLastSeen = seen
}

// LoggedIn handles everything that needs to happen when a Character enters the game.
func (c *Character) LoggedIn() {
	room := c.Room()
	area := c.Room().ParentArea

	// Add character to room
	if room == nil || area == nil {
		Armeria.log.Fatal("character logged into an invalid area/room",
			zap.String("character", c.Name()),
		)
		return
	}

	// Show server / character info
	c.Player().client.ShowText(
		fmt.Sprintf(
			"The server has been running for %s.\n"+
				"You last logged in at %s (server time). ",
			TextStyle(time.Since(Armeria.startTime), WithBold()),
			TextStyle(c.LastSeen().Format("Mon Jan 2 2006 15:04:05 MST"), WithBold()),
		),
	)

	// Update lastSeen
	c.SetLastSeen(time.Now())

	// Use command: /look
	Armeria.commandManager.ProcessCommand(c.Player(), "look", false)

	// Show message to others in the same room
	for _, char := range room.Here().Characters(true, c) {
		pc := char.Player()
		pc.client.ShowText(
			fmt.Sprintf("%s connected and appeared here with you.", c.Name()),
		)
	}

	area.CharacterEntered(c, true)
	room.CharacterEntered(c, true)

	c.Player().client.SyncInventory()
	c.Player().client.SyncPermissions()
	c.Player().client.SyncPlayerInfo()
	c.Player().client.SyncMoney()
	c.Player().client.SyncCommands()
	c.Player().client.SyncSettings()

	Armeria.log.Info("character entered the game",
		zap.String("character", c.Name()),
	)
}

// LoggedOut handles everything that needs to happen when a Character leaves the game.
func (c *Character) LoggedOut() {
	room := c.Room()
	area := c.Room().ParentArea

	// Remove unsafeCharacter from room
	if room == nil || area == nil {
		Armeria.log.Fatal("character logged out of an invalid area/room",
			zap.String("character", c.Name()),
		)
		return
	}

	// Show message to others in the same room
	for _, char := range room.Here().Characters(true, c) {
		pc := char.Player()
		pc.client.ShowText(
			fmt.Sprintf("%s disconnected and is no longer here with you.", c.Name()),
		)
	}

	area.CharacterLeft(c, true)
	room.CharacterLeft(c, true)

	// Clear temp attributes
	for key := range c.UnsafeTempAttributes {
		delete(c.UnsafeTempAttributes, key)
	}

	// Stop any on-going mob conversations
	if c.MobConvo() != nil {
		c.MobConvo().Cancel()
	}

	Armeria.log.Info("character left the game",
		zap.String("character", c.Name()),
	)
}

// TempAttribute retrieves a previously-saved temp attribute.
func (c *Character) TempAttribute(name string) string {
	c.RLock()
	defer c.RUnlock()

	return c.UnsafeTempAttributes[name]
}

// SetTempAttribute sets a temporary attribute, which is cleared on log out. Additionally, these
// attributes are not validated.
func (c *Character) SetTempAttribute(name string, value string) {
	c.Lock()
	defer c.Unlock()

	if c.UnsafeTempAttributes == nil {
		c.UnsafeTempAttributes = make(map[string]string)
	}

	c.UnsafeTempAttributes[name] = value
}

// SetAttribute sets a permanent attribute and only valid attributes can be set.
func (c *Character) SetAttribute(name string, value string) error {
	c.Lock()
	defer c.Unlock()

	if !misc.Contains(AttributeList(ObjectTypeCharacter), name) {
		return errors.New("attribute name is invalid")
	}

	c.UnsafeAttributes[name] = value
	return nil
}

// Attribute returns a permanent attribute.
func (c *Character) Attribute(name string) string {
	c.RLock()
	defer c.RUnlock()

	if len(c.UnsafeAttributes[name]) == 0 {
		return AttributeDefault(ObjectTypeCharacter, name)
	}

	return c.UnsafeAttributes[name]
}

// Money returns the character's money as a float.
func (c *Character) Money() float64 {
	money := c.Attribute(AttributeMoney)
	f, err := strconv.ParseFloat(money, 64)
	if err != nil {
		Armeria.log.Fatal("unable to convert money to float64",
			zap.Error(err),
		)
	}
	return f
}

// RemoveMoney attempts to remove money from the character and returns True if they can afford it.
func (c *Character) RemoveMoney(amount float64) bool {
	money := c.Money()
	if amount > money {
		return false
	}

	_ = c.SetAttribute(AttributeMoney, fmt.Sprintf("%.2f", money-amount))

	return true
}

// AddMoney adds money to the character.
func (c *Character) AddMoney(amount float64) {
	_ = c.SetAttribute(AttributeMoney, fmt.Sprintf("%.2f", c.Money()+amount))
}

// SetSetting sets a Character setting and only valid settings can be set.
func (c *Character) SetSetting(name string, value string) error {
	c.Lock()
	defer c.Unlock()

	if !misc.Contains(ValidSettings(), name) {
		return errors.New("setting name is invalid")
	}
	c.UnsafeSettings[name] = value
	return nil
}

// Setting returns a setting's value.
func (c *Character) Setting(name string) string {
	c.RLock()
	defer c.RUnlock()

	if len(c.UnsafeSettings[name]) == 0 {
		return SettingDefault(name)
	}

	return c.UnsafeSettings[name]
}

// MoveAllowed will check if moving to a particular location is valid/allowed.
func (c *Character) MoveAllowed(r *Room) (bool, string) {
	if r == nil {
		return false, CommonInvalidDirection
	}

	if len(c.TempAttribute(TempAttributeGhost)) > 0 {
		return true, ""
	}

	if r.Attribute("type") == "track" {
		return false, "You cannot walk onto the train tracks!"
	}

	return true, ""
}

// Move will move the Character to a new location (no move checks are performed).
func (c *Character) Move(to *Room, msgToChar string, msgToOld string, msgToNew string, sfx sfx.ClientSoundEffect) {
	oldRoom := c.Room()
	if oldRoom == nil {
		// If the character logged out in a room that no longer exists, allow movement to still work so they
		// can be teleported by a staff member to a "real" room.
		oldRoom = to
	}

	oldRoom.Here().Remove(c.ID())
	if err := to.Here().Add(c.ID()); err != nil {
		Armeria.log.Fatal("error adding character to destination room")
	}

	if !c.Online() {
		return
	}

	for _, char := range oldRoom.Here().Characters(true) {
		char.Player().client.ShowText(msgToOld)
		if len(sfx) > 0 {
			char.Player().client.PlaySFX(sfx)
		}
	}

	for _, char := range to.Here().Characters(true, c) {
		char.Player().client.ShowText(msgToNew)
		if len(sfx) > 0 {
			char.Player().client.PlaySFX(sfx)
		}
	}

	c.Player().client.ShowText(msgToChar)
	if len(sfx) > 0 {
		c.Player().client.PlaySFX(sfx)
	}

	oldArea := oldRoom.ParentArea
	newArea := to.ParentArea
	if oldArea.ID() != newArea.ID() {
		oldArea.CharacterLeft(c, false)
		newArea.CharacterEntered(c, false)
	}

	oldRoom.CharacterEntered(c, false)
	to.CharacterEntered(c, false)

	// Stop any on-going mob conversations.
	if c.MobConvo() != nil {
		c.MobConvo().Cancel()
	}

	// If the object editor is open, move the editor to this room.
	if c.TempAttribute(TempAttributeEditorOpen) == "true" {
		c.Player().client.ShowObjectEditor(to.EditorData())
	}
}

// EditorData returns the JSON used for the object editor.
func (c *Character) EditorData() *ObjectEditorData {
	var props []*ObjectEditorDataProperty
	for _, attrName := range AttributeList(ObjectTypeCharacter) {
		props = append(props, &ObjectEditorDataProperty{
			PropType: AttributeEditorType(ObjectTypeCharacter, attrName),
			Name:     attrName,
			Group:    AttributeGroup(attrName),
			Value:    c.Attribute(attrName),
		})
	}

	return &ObjectEditorData{
		UUID:       c.ID(),
		Name:       c.Name(),
		ObjectType: "character",
		Properties: props,
	}
}

func (c *Character) SettingsJSON() string {
	c.RLock()
	defer c.RUnlock()

	obj := make(map[string]string)
	for _, setting := range ValidSettings() {
		s, exists := c.UnsafeSettings[setting]
		if exists {
			obj[setting] = s
		} else {
			obj[setting] = SettingDefault(setting)
		}
	}

	b, err := json.Marshal(obj)
	if err != nil {
		log.Fatalf("error marshalling character settings: %s", err)
	}

	return string(b)
}

// HasPermission returns true if the Character has a particular permission.
func (c *Character) HasPermission(p string) bool {
	c.RLock()
	defer c.RUnlock()

	perms := strings.Split(c.UnsafeAttributes[AttributePermissions], " ")
	return misc.Contains(perms, p)
}

// Channels returns the Channel objects for the channels this unsafeCharacter is within.
func (c *Character) Channels() []*Channel {
	var channels []*Channel

	for _, channel := range strings.Split(c.Attribute(AttributeChannels), ",") {
		ch := ChannelByName(channel)
		if ch != nil {
			channels = append(channels, ch)
		}
	}

	return channels
}

// InChannel returns true if the Character is in a particular channel.
func (c *Character) InChannel(ch *Channel) bool {
	c.RLock()
	defer c.RUnlock()

	channelsString := c.UnsafeAttributes[AttributeChannels]
	return misc.Contains(strings.Split(strings.ToLower(channelsString), ","), strings.ToLower(ch.Name))
}

// Online is used to see if the character is online.
func (c *Character) Online() bool {
	return c.Player() != nil
}

// JoinChannel adds a channel to the Character's channel list so that they will receive messages
// on that channel.
func (c *Character) JoinChannel(ch *Channel) {
	chs := strings.Split(c.Attribute(AttributeChannels), ",")
	if len(chs[0]) == 0 {
		chs[0] = ch.Name
	} else {
		chs = append(chs, ch.Name)
	}
	_ = c.SetAttribute(AttributeChannels, strings.Join(chs, ","))
}

// LeaveChannel removes the channel from the Character's channel list.
func (c *Character) LeaveChannel(ch *Channel) {
	chs := strings.Split(c.Attribute(AttributeChannels), ",")
	for i, cname := range chs {
		if cname == ch.Name {
			chs[i] = chs[len(chs)-1]
			chs = chs[:len(chs)-1]
			break
		}
	}
	_ = c.SetAttribute(AttributeChannels, strings.Join(chs, ","))
}

// InventoryJSON returns the JSON used for rendering the inventory on the client.
func (c *Character) InventoryJSON() string {
	var inventory []map[string]interface{}

	for _, ii := range c.Inventory().Items() {
		inventory = append(inventory, map[string]interface{}{
			"uuid":      ii.ID(),
			"name":      ii.Name(),
			"picture":   ii.Attribute(AttributePicture),
			"slot":      c.Inventory().Slot(ii.ID()),
			"equipSlot": ii.Attribute(AttributeEquipSlot),
			"color":     ii.RarityColor(),
		})
	}

	inventoryJSON, err := json.Marshal(inventory)
	if err != nil {
		Armeria.log.Fatal("failed to marshal inventory data",
			zap.String("character", c.UUID),
			zap.Error(err),
		)
	}

	return string(inventoryJSON)
}

// Pronoun is used to determine the appropriate pronoun for the character.
func (c *Character) Pronoun(pt PronounType) string {
	gender := c.Attribute(AttributeGender)
	if gender == "male" {
		if pt == PronounSubjective {
			return "he"
		} else if pt == PronounPossessiveAbsolute {
			return "his"
		} else if pt == PronounPossessiveAdjective {
			return "his"
		} else if pt == PronounObjective {
			return "him"
		}
	} else if gender == "female" {
		if pt == PronounSubjective {
			return "she"
		} else if pt == PronounPossessiveAbsolute {
			return "hers"
		} else if pt == PronounPossessiveAdjective {
			return "her"
		} else if pt == PronounObjective {
			return "her"
		}
	}

	return ""
}
