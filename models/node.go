package models

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gravitl/netmaker/database"
	"golang.org/x/crypto/bcrypt"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const TEN_YEARS_IN_SECONDS = 300000000

// == ACTIONS == (can only be set by GRPC)
const NODE_UPDATE_KEY = "updatekey"
const NODE_DELETE = "delete"
const NODE_IS_PENDING = "pending"
const NODE_NOOP = "noop"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//node struct
type Node struct {
	ID                  string   `json:"id,omitempty" bson:"id,omitempty"`
	Address             string   `json:"address" bson:"address" yaml:"address" validate:"omitempty,ipv4"`
	Address6            string   `json:"address6" bson:"address6" yaml:"address6" validate:"omitempty,ipv6"`
	LocalAddress        string   `json:"localaddress" bson:"localaddress" yaml:"localaddress" validate:"omitempty,ip"`
	Name                string   `json:"name" bson:"name" yaml:"name" validate:"omitempty,max=12,in_charset"`
	ListenPort          int32    `json:"listenport" bson:"listenport" yaml:"listenport" validate:"omitempty,numeric,min=1024,max=65535"`
	PublicKey           string   `json:"publickey" bson:"publickey" yaml:"publickey" validate:"required,base64"`
	Endpoint            string   `json:"endpoint" bson:"endpoint" yaml:"endpoint" validate:"required,ip"`
	PostUp              string   `json:"postup" bson:"postup" yaml:"postup"`
	PostDown            string   `json:"postdown" bson:"postdown" yaml:"postdown"`
	AllowedIPs          []string `json:"allowedips" bson:"allowedips" yaml:"allowedips"`
	PersistentKeepalive int32    `json:"persistentkeepalive" bson:"persistentkeepalive" yaml:"persistentkeepalive" validate:"omitempty,numeric,max=1000"`
	SaveConfig          string   `json:"saveconfig" bson:"saveconfig" yaml:"saveconfig" validate:"checkyesorno"`
	AccessKey           string   `json:"accesskey" bson:"accesskey" yaml:"accesskey"`
	Interface           string   `json:"interface" bson:"interface" yaml:"interface"`
	LastModified        int64    `json:"lastmodified" bson:"lastmodified" yaml:"lastmodified"`
	KeyUpdateTimeStamp  int64    `json:"keyupdatetimestamp" bson:"keyupdatetimestamp" yaml:"keyupdatetimestamp"`
	ExpirationDateTime  int64    `json:"expdatetime" bson:"expdatetime" yaml:"expdatetime"`
	LastPeerUpdate      int64    `json:"lastpeerupdate" bson:"lastpeerupdate" yaml:"lastpeerupdate"`
	LastCheckIn         int64    `json:"lastcheckin" bson:"lastcheckin" yaml:"lastcheckin"`
	MacAddress          string   `json:"macaddress" bson:"macaddress" yaml:"macaddress" validate:"required,mac,macaddress_unique"`
	CheckInInterval     int32    `json:"checkininterval" bson:"checkininterval" yaml:"checkininterval"`
	Password            string   `json:"password" bson:"password" yaml:"password" validate:"required,min=6"`
	Network             string   `json:"network" bson:"network" yaml:"network" validate:"network_exists"`
	IsPending           string   `json:"ispending" bson:"ispending" yaml:"ispending"`
	IsEgressGateway     string   `json:"isegressgateway" bson:"isegressgateway" yaml:"isegressgateway"`
	IsIngressGateway    string   `json:"isingressgateway" bson:"isingressgateway" yaml:"isingressgateway"`
	EgressGatewayRanges []string `json:"egressgatewayranges" bson:"egressgatewayranges" yaml:"egressgatewayranges"`
	IngressGatewayRange string   `json:"ingressgatewayrange" bson:"ingressgatewayrange" yaml:"ingressgatewayrange"`
	StaticIP            string   `json:"staticip" bson:"staticip" yaml:"staticip"`
	StaticPubKey        string   `json:"staticpubkey" bson:"staticpubkey" yaml:"staticpubkey"`
	UDPHolePunch        string   `json:"udpholepunch" bson:"udpholepunch" yaml:"udpholepunch" validate:"checkyesorno"`
	PullChanges         string   `json:"pullchanges" bson:"pullchanges" yaml:"pullchanges" validate:"checkyesorno"`
	DNSOn               string   `json:"dnson" bson:"dnson" yaml:"dnson" validate:"checkyesorno"`
	IsDualStack         string   `json:"isdualstack" bson:"isdualstack" yaml:"isdualstack" validate:"checkyesorno"`
	Action              string   `json:"action" bson:"action" yaml:"action"`
	IsLocal             string   `json:"islocal" bson:"islocal" yaml:"islocal" validate:"checkyesorno"`
	LocalRange          string   `json:"localrange" bson:"localrange" yaml:"localrange"`
	Roaming             string   `json:"roaming" bson:"roaming" yaml:"roaming" validate:"checkyesorno"`
	IPForwarding        string   `json:"ipforwarding" bson:"ipforwarding" yaml:"ipforwarding" validate:"checkyesorno"`
}

func (node *Node) SetDefaultAction() {
	if node.Action == "" {
		node.Action = NODE_NOOP
	}
}

func (node *Node) SetRoamingDefault() {
	if node.Roaming == "" {
		node.Roaming = "no"
	}
}

func (node *Node) SetPullChangesDefault() {
	if node.PullChanges == "" {
		node.PullChanges = "no"
	}
}

func (node *Node) SetIPForwardingDefault() {
	if node.IPForwarding == "" {
		node.IPForwarding = "no"
	}
}

func (node *Node) SetIsLocalDefault() {
	if node.IsLocal == "" {
		node.IsLocal = "no"
	}
}

func (node *Node) SetDNSOnDefault() {
	if node.DNSOn == "" {
		node.DNSOn = "no"
	}
}

func (node *Node) SetIsDualStackDefault() {
	if node.IsDualStack == "" {
		node.IsDualStack = "no"
	}
}

func (node *Node) SetLastModified() {
	node.LastModified = time.Now().Unix()
}

func (node *Node) SetLastCheckIn() {
	node.LastCheckIn = time.Now().Unix()
}

func (node *Node) SetLastPeerUpdate() {
	node.LastPeerUpdate = time.Now().Unix()
}

func (node *Node) SetID() {
	node.ID = node.MacAddress + "###" + node.Network
}

func (node *Node) SetExpirationDateTime() {
	node.ExpirationDateTime = time.Now().Unix() + TEN_YEARS_IN_SECONDS
}

func (node *Node) SetDefaultName() {
	if node.Name == "" {
		nodeid := StringWithCharset(5, charset)
		nodename := "node-" + nodeid
		node.Name = nodename
	}
}

func (node *Node) GetNetwork() (Network, error) {

	var network Network
	networkData, err := database.FetchRecord(database.NETWORKS_TABLE_NAME, node.Network)
	if err != nil {
		return network, err
	}
	if err = json.Unmarshal([]byte(networkData), &network); err != nil {
		return Network{}, err
	}
	return network, nil
}

//TODO: I dont know why this exists
//This should exist on the node.go struct. I'm sure there was a reason?
func (node *Node) SetDefaults() {

	//TODO: Maybe I should make Network a part of the node struct. Then we can just query the Network object for stuff.
	parentNetwork, _ := node.GetNetwork()

	node.ExpirationDateTime = time.Now().Unix() + TEN_YEARS_IN_SECONDS

	if node.ListenPort == 0 {
		node.ListenPort = parentNetwork.DefaultListenPort
	}
	if node.SaveConfig == "" {
		if parentNetwork.DefaultSaveConfig != "" {
			node.SaveConfig = parentNetwork.DefaultSaveConfig
		} else {
			node.SaveConfig = "yes"
		}
	}
	if node.Interface == "" {
		node.Interface = parentNetwork.DefaultInterface
	}
	if node.PersistentKeepalive == 0 {
		node.PersistentKeepalive = parentNetwork.DefaultKeepalive
	}
	if node.PostUp == "" {
		postup := parentNetwork.DefaultPostUp
		node.PostUp = postup
	}
	if node.StaticIP == "" {
		node.StaticIP = "no"
	}
	if node.StaticPubKey == "" {
		node.StaticPubKey = "no"
	}
	if node.UDPHolePunch == "" {
		node.UDPHolePunch = parentNetwork.DefaultUDPHolePunch
		if node.UDPHolePunch == "" {
			node.UDPHolePunch = "yes"
		}
	}
	node.CheckInInterval = parentNetwork.DefaultCheckInInterval
	node.SetIPForwardingDefault()
	node.SetDNSOnDefault()
	node.SetIsLocalDefault()
	node.SetIsDualStackDefault()
	node.SetLastModified()
	node.SetDefaultName()
	node.SetLastCheckIn()
	node.SetLastPeerUpdate()
	node.SetRoamingDefault()
	node.SetPullChangesDefault()
	node.SetDefaultAction()
	node.SetID()
	node.KeyUpdateTimeStamp = time.Now().Unix()
}

func (newNode *Node) Fill(currentNode *Node) {
	if newNode.ID == "" {
		newNode.ID = currentNode.ID
	}
	if newNode.Address == "" {
		newNode.Address = currentNode.Address
	}
	if newNode.Address6 == "" {
		newNode.Address6 = currentNode.Address6
	}
	if newNode.LocalAddress == "" {
		newNode.LocalAddress = currentNode.LocalAddress
	}
	if newNode.Name == "" {
		newNode.Name = currentNode.Name
	}
	if newNode.ListenPort == 0 {
		newNode.ListenPort = currentNode.ListenPort
	}
	if newNode.PublicKey == "" {
		newNode.PublicKey = currentNode.PublicKey
	} else {
		newNode.KeyUpdateTimeStamp = time.Now().Unix()
	}
	if newNode.Endpoint == "" {
		newNode.Endpoint = currentNode.Endpoint
	}
	if newNode.PostUp == "" {
		newNode.PostUp = currentNode.PostUp
	}
	if newNode.PostDown == "" {
		newNode.PostDown = currentNode.PostDown
	}
	if newNode.AllowedIPs == nil {
		newNode.AllowedIPs = currentNode.AllowedIPs
	}
	if newNode.PersistentKeepalive == 0 {
		newNode.PersistentKeepalive = currentNode.PersistentKeepalive
	}
	if newNode.SaveConfig == "" {
		newNode.SaveConfig = currentNode.SaveConfig
	}
	if newNode.AccessKey == "" {
		newNode.AccessKey = currentNode.AccessKey
	}
	if newNode.Interface == "" {
		newNode.Interface = currentNode.Interface
	}
	if newNode.LastModified == 0 {
		newNode.LastModified = currentNode.LastModified
	}
	if newNode.KeyUpdateTimeStamp == 0 {
		newNode.LastModified = currentNode.LastModified
	}
	if newNode.ExpirationDateTime == 0 {
		newNode.ExpirationDateTime = currentNode.ExpirationDateTime
	}
	if newNode.LastPeerUpdate == 0 {
		newNode.LastPeerUpdate = currentNode.LastPeerUpdate
	}
	if newNode.LastCheckIn == 0 {
		newNode.LastCheckIn = currentNode.LastCheckIn
	}
	if newNode.MacAddress == "" {
		newNode.MacAddress = currentNode.MacAddress
	}
	if newNode.CheckInInterval == 0 {
		newNode.CheckInInterval = currentNode.CheckInInterval
	}
	if newNode.Password != "" {
		err := bcrypt.CompareHashAndPassword([]byte(newNode.Password), []byte(currentNode.Password))
		if err != nil && currentNode.Password != newNode.Password {
			hash, err := bcrypt.GenerateFromPassword([]byte(newNode.Password), 5)
			if err == nil {
				newNode.Password = string(hash)
			}
		}
	} else {
		newNode.Password = currentNode.Password
	}
	if newNode.Network == "" {
		newNode.Network = currentNode.Network
	}
	if newNode.IsPending == "" {
		newNode.IsPending = currentNode.IsPending
	}
	if newNode.IsEgressGateway == "" {
		newNode.IsEgressGateway = currentNode.IsEgressGateway
	}
	if newNode.IsIngressGateway == "" {
		newNode.IsIngressGateway = currentNode.IsIngressGateway
	}
	if newNode.EgressGatewayRanges == nil {
		newNode.EgressGatewayRanges = currentNode.EgressGatewayRanges
	}
	if newNode.IngressGatewayRange == "" {
		newNode.IngressGatewayRange = currentNode.IngressGatewayRange
	}
	if newNode.StaticIP == "" {
		newNode.StaticIP = currentNode.StaticIP
	}
	if newNode.StaticIP == "" {
		newNode.StaticIP = currentNode.StaticIP
	}
	if newNode.StaticPubKey == "" {
		newNode.StaticPubKey = currentNode.StaticPubKey
	}
	if newNode.UDPHolePunch == "" {
		newNode.UDPHolePunch = currentNode.SaveConfig
	}
	if newNode.DNSOn == "" {
		newNode.DNSOn = currentNode.DNSOn
	}
	if newNode.IsDualStack == "" {
		newNode.IsDualStack = currentNode.IsDualStack
	}
	if newNode.IsLocal == "" {
		newNode.IsLocal = currentNode.IsLocal
	}
	if newNode.IPForwarding == "" {
		newNode.IPForwarding = currentNode.IPForwarding
	}
	if newNode.PullChanges == "" {
		newNode.PullChanges = currentNode.PullChanges
	}
	if newNode.Roaming == "" {
		newNode.Roaming = currentNode.Roaming
	}
	if newNode.Action == "" {
		newNode.Action = currentNode.Action
	}
}

func (currentNode *Node) Update(newNode *Node) error {
	newNode.Fill(currentNode)
	if err := newNode.Validate(true); err != nil {
		return err
	}
	newNode.SetID()
	if newNode.ID == currentNode.ID {
		newNode.SetLastModified()
		if data, err := json.Marshal(newNode); err != nil {
			return err
		} else {
			if err = database.Insert(newNode.ID, string(data), database.NODES_TABLE_NAME); err == nil {
				if network, err := GetNetwork(newNode.Network); err == nil {
					err = network.SetNetworkNodesLastModified()
				}
			}
			return err
		}
	}
	return errors.New("failed to update node " + newNode.MacAddress + ", cannot change macaddress.")
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

//Check for valid IPv4 address
//Note: We dont handle IPv6 AT ALL!!!!! This definitely is needed at some point
//But for iteration 1, lets just stick to IPv4. Keep it simple stupid.
func IsIpv4Net(host string) bool {
	return net.ParseIP(host) != nil
}

func (node *Node) Validate(isUpdate bool) error {
	v := validator.New()
	_ = v.RegisterValidation("macaddress_unique", func(fl validator.FieldLevel) bool {
		if isUpdate {
			return true
		}
		isFieldUnique, _ := node.IsIDUnique()
		return isFieldUnique
	})
	_ = v.RegisterValidation("network_exists", func(fl validator.FieldLevel) bool {
		_, err := node.GetNetwork()
		return err == nil
	})
	_ = v.RegisterValidation("in_charset", func(fl validator.FieldLevel) bool {
		isgood := node.NameInNodeCharSet()
		return isgood
	})
	_ = v.RegisterValidation("checkyesorno", func(fl validator.FieldLevel) bool {
		return CheckYesOrNo(fl)
	})
	err := v.Struct(node)

	return err
}

func (node *Node) IsIDUnique() (bool, error) {
	_, err := database.FetchRecord(database.NODES_TABLE_NAME, node.ID)
	return database.IsEmptyRecord(err), err
}

func (node *Node) NameInNodeCharSet() bool {

	charset := "abcdefghijklmnopqrstuvwxyz1234567890-"

	for _, char := range node.Name {
		if !strings.Contains(charset, strings.ToLower(string(char))) {
			return false
		}
	}
	return true
}

func GetAllNodes() ([]Node, error) {
	var nodes []Node

	collection, err := database.FetchRecords(database.NODES_TABLE_NAME)
	if err != nil {
		if database.IsEmptyRecord(err) {
			return []Node{}, nil
		}
		return []Node{}, err
	}

	for _, value := range collection {
		var node Node
		if err := json.Unmarshal([]byte(value), &node); err != nil {
			return []Node{}, err
		}
		// add node to our array
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func GetNode(macaddress string, network string) (Node, error) {

	var node Node

	key, err := GetID(macaddress, network)
	if err != nil {
		return node, err
	}
	data, err := database.FetchRecord(database.NODES_TABLE_NAME, key)
	if err != nil {
		return node, err
	}
	if err = json.Unmarshal([]byte(data), &node); err != nil {
		return node, err
	}

	return node, err
}

func GetID(macaddress string, network string) (string, error) {
	if macaddress == "" || network == "" {
		return "", errors.New("unable to get record key")
	}
	return macaddress + "###" + network, nil
}
