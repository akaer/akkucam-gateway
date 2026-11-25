package ppcs

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

// HubInfo is the verified metadata reported by the base station.
type HubInfo struct {
	Model        string `json:"model,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	Firmware     string `json:"firmware,omitempty"`
	APIVersion   string `json:"apiVersion,omitempty"`
	MCUVersion   string `json:"mcuVersion,omitempty"`
	MAC          string `json:"mac,omitempty"`
	SDCardStatus string `json:"sdCardStatus,omitempty"`
	SDTotalBytes *int64 `json:"sdTotalBytes,omitempty"`
	SDFreeBytes  *int64 `json:"sdFreeBytes,omitempty"`
}

// CameraChannel is one coupled camera as reported by the base station.
type CameraChannel struct {
	Index             int    `json:"index"`
	Name              string `json:"name,omitempty"`
	Model             string `json:"model,omitempty"`
	FWVersion         string `json:"fwVersion,omitempty"`
	MCUVersion        string `json:"mcuVersion,omitempty"`
	FWVersionRaw      uint32 `json:"fwVersionRaw,omitempty"`
	MCUVersionRaw     uint32 `json:"mcuVersionRaw,omitempty"`
	Updateable        *bool  `json:"updateable,omitempty"`
	MotionEnabled     *bool  `json:"motionEnabled,omitempty"`
	PIRMode           *int   `json:"pirMode,omitempty"`
	PIRSensitivityDay *int   `json:"pirSensitivityDay,omitempty"`
}

// Inventory is the hub metadata plus all coupled camera channels.
type Inventory struct {
	Hub      HubInfo         `json:"hub"`
	Channels []CameraChannel `json:"channels"`
}

type controlClient struct {
	live     *Live
	receiver *channelReceiver
	buffer   []byte
}

// QueryInventory authenticates locally and reads hub metadata plus the coupled
// camera list.
func QueryInventory(ctx context.Context, ip, did, password string) (Inventory, error) {
	return queryInventoryAt(ctx, ip, did, password, &net.UDPAddr{IP: net.IPv4bcast, Port: LanSearchPort})
}

// QueryChannels remains a narrow convenience API for protocol users.
func QueryChannels(ctx context.Context, ip, did, password string) ([]CameraChannel, error) {
	inventory, err := QueryInventory(ctx, ip, did, password)
	return inventory.Channels, err
}

func queryChannelsAt(ctx context.Context, ip, did, password string, searchTarget *net.UDPAddr) ([]CameraChannel, error) {
	inventory, err := queryInventoryAt(ctx, ip, did, password, searchTarget)
	return inventory.Channels, err
}

func queryInventoryAt(ctx context.Context, ip, did, password string, searchTarget *net.UDPAddr) (Inventory, error) {
	session, key, _, err := openAuthenticatedAt(ctx, ip, did, password, searchTarget)
	if err != nil {
		return Inventory{}, err
	}
	live := &Live{session: session, key: key, xorCtl: controlInitialXor, rx: make(map[byte]*channelReceiver)}
	defer live.Close()
	client := &controlClient{live: live, receiver: &channelReceiver{assembler: &appAssembler{}}, buffer: make([]byte, liveReadBuffer)}

	listPayload, err := client.request(ctx, getCameraListControl, make([]byte, controlHeaderSize), getCameraListControl)
	if err != nil {
		return Inventory{}, err
	}
	channels, err := parseCameraList(listPayload)
	if err != nil {
		return Inventory{}, err
	}

	inventory := Inventory{Channels: channels}
	if payload, infoErr := client.request(ctx, getDeviceInfoControl, make([]byte, deviceInfoRequestSize), deviceInfoResponseControl); infoErr == nil {
		inventory.Hub, _ = parseHubInfo(payload)
	}
	for index := range inventory.Channels {
		request := make([]byte, controlHeaderSize)
		request[motionChannelOffset] = byte(inventory.Channels[index].Index)
		payload, paramErr := client.request(ctx, getCameraParametersControl, request, getCameraParametersControl)
		if paramErr != nil {
			continue
		}
		parameters, parseErr := parseCameraParameters(payload)
		if parseErr != nil || parameters[motionChannelOffset] != byte(inventory.Channels[index].Index) {
			continue
		}
		mode, sensitivity := int(parameters[motionModeOffset]), int(parameters[motionSensitivityOff])
		enabled := mode != motionDisabledValue
		inventory.Channels[index].PIRMode = &mode
		inventory.Channels[index].PIRSensitivityDay = &sensitivity
		inventory.Channels[index].MotionEnabled = &enabled
	}
	return inventory, nil
}

// SetMotion changes the PIR motion mode using read-modify-write-verify. The
// full 256-byte parameter block is written back unchanged except for pirMode.
func SetMotion(ctx context.Context, ip, did, password string, channel int, enabled bool) (CameraChannel, error) {
	return setMotionAt(ctx, ip, did, password, channel, enabled, &net.UDPAddr{IP: net.IPv4bcast, Port: LanSearchPort})
}

func setMotionAt(ctx context.Context, ip, did, password string, channel int, enabled bool, searchTarget *net.UDPAddr) (CameraChannel, error) {
	if channel < 0 || channel > MaxMotionChannel {
		return CameraChannel{}, ErrChannelRange
	}
	session, key, _, err := openAuthenticatedAt(ctx, ip, did, password, searchTarget)
	if err != nil {
		return CameraChannel{}, err
	}
	live := &Live{session: session, key: key, xorCtl: controlInitialXor, rx: make(map[byte]*channelReceiver)}
	defer live.Close()
	client := &controlClient{live: live, receiver: &channelReceiver{assembler: &appAssembler{}}, buffer: make([]byte, liveReadBuffer)}
	request := make([]byte, controlHeaderSize)
	request[motionChannelOffset] = byte(channel)
	parameters, err := client.request(ctx, getCameraParametersControl, request, getCameraParametersControl)
	if err != nil {
		return CameraChannel{}, err
	}
	parameters, err = parseCameraParameters(parameters)
	if err != nil {
		return CameraChannel{}, err
	}
	if int(parameters[motionChannelOffset]) != channel {
		return CameraChannel{}, fmt.Errorf("%w: response belongs to channel %d", ErrCameraParameters, parameters[motionChannelOffset])
	}
	if enabled {
		parameters[motionModeOffset] = motionEnabledValue
	} else {
		parameters[motionModeOffset] = motionDisabledValue
	}
	if err := live.sendControl(setCameraParametersControl, parameters); err != nil {
		return CameraChannel{}, err
	}

	// Verify with a fresh read. This avoids reporting success on a dropped write.
	verified, err := client.request(ctx, getCameraParametersControl, request, getCameraParametersControl)
	if err != nil {
		return CameraChannel{}, err
	}
	verified, err = parseCameraParameters(verified)
	if err != nil {
		return CameraChannel{}, err
	}
	mode, sensitivity := int(verified[motionModeOffset]), int(verified[motionSensitivityOff])
	actual := mode != motionDisabledValue
	if actual != enabled {
		return CameraChannel{}, ErrMotionNotApplied
	}
	return CameraChannel{Index: channel, MotionEnabled: &actual, PIRMode: &mode, PIRSensitivityDay: &sensitivity}, nil
}

func (c *controlClient) request(ctx context.Context, requestType uint16, requestPayload []byte, responseType uint16) ([]byte, error) {
	if err := c.live.sendControl(requestType, requestPayload); err != nil {
		return nil, err
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("IO-Control %d timeout: %w", responseType, err)
		}
		_ = c.live.session.conn.SetReadDeadline(shortDeadline(ctx, sessionReadPoll))
		n, remote, err := c.live.session.conn.ReadFromUDP(c.buffer)
		if err != nil {
			var netErr net.Error
			if errorsAsTimeout(err, &netErr) {
				continue
			}
			return nil, err
		}
		if remote.String() != c.live.session.peer.String() || n < ppcsHeaderSize {
			continue
		}
		header, err := ParseHeader(c.buffer[:n])
		if err != nil {
			continue
		}
		switch header.Type {
		case MsgAlive:
			_ = c.live.session.send(Packet(MsgAliveAck, nil))
		case msgDRW:
			channel, sequence, chunk, err := parseDRW(c.buffer[:n])
			if err != nil {
				continue
			}
			_ = c.live.session.send(buildDRWAck(channel, []uint16{sequence}))
			if channel != controlChannel {
				continue
			}
			for _, ordered := range c.receiver.add(sequence, chunk) {
				for _, frame := range c.receiver.assembler.feed(ordered) {
					controlType, payload, ok := c.live.decodeControl(frame)
					if ok && controlType == responseType {
						return payload, nil
					}
				}
			}
		}
	}
}

func errorsAsTimeout(err error, target *net.Error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		*target = netErr
		return netErr.Timeout()
	}
	return false
}

func parseCameraList(payload []byte) ([]CameraChannel, error) {
	if len(payload) < cameraListCountSize {
		return nil, fmt.Errorf("%w: shorter than count field", ErrCameraListTruncated)
	}
	count := int(binary.LittleEndian.Uint32(payload[0:cameraListCountSize]))
	if count > cameraListMax {
		return nil, fmt.Errorf("%w: %d", ErrCameraListCount, count)
	}
	required := cameraListCountSize + count*cameraListEntrySize
	if len(payload) < required {
		return nil, fmt.Errorf("%w: need %d bytes, got %d", ErrCameraListTruncated, required, len(payload))
	}
	channels := make([]CameraChannel, 0, count)
	for index := 0; index < count; index++ {
		offset := cameraListCountSize + index*cameraListEntrySize
		fwRaw := binary.LittleEndian.Uint32(payload[offset+cameraListFWOffset : offset+cameraListFWOffset+cameraListVersionSize])
		mcuRaw := binary.LittleEndian.Uint32(payload[offset+cameraListMCUOffset : offset+cameraListMCUOffset+cameraListVersionSize])
		updateable := payload[offset+cameraListUpdateOffset] == cameraListUpdateableValue
		channels = append(channels, CameraChannel{
			Index:         int(payload[offset+cameraListIndexOffset]),
			Name:          cleanFixedString(payload[offset+cameraListNameOffset : offset+cameraListNameOffset+cameraListNameSize]),
			Model:         cleanFixedString(payload[offset+cameraListModelOffset : offset+cameraListModelOffset+cameraListModelSize]),
			FWVersion:     versionBigIfPresent(fwRaw),
			MCUVersion:    versionBigIfPresent(mcuRaw),
			FWVersionRaw:  fwRaw,
			MCUVersionRaw: mcuRaw,
			Updateable:    &updateable,
		})
	}
	return channels, nil
}

func parseCameraParameters(payload []byte) ([]byte, error) {
	if len(payload) < cameraParametersSize {
		return nil, fmt.Errorf("%w: got %d bytes", ErrCameraParameters, len(payload))
	}
	return append([]byte(nil), payload[:cameraParametersSize]...), nil
}

func parseHubInfo(payload []byte) (HubInfo, error) {
	if len(payload) < hubInfoMinSize {
		return HubInfo{}, fmt.Errorf("%w: got %d bytes", ErrHubInfoTruncated, len(payload))
	}
	info := HubInfo{
		Model:      cleanFixedString(payload[hubInfoModelOffset : hubInfoModelOffset+hubInfoModelSize]),
		Vendor:     cleanFixedString(payload[hubInfoVendorOffset : hubInfoVendorOffset+hubInfoVendorSize]),
		Firmware:   versionReversedIfPresent(payload[hubInfoVersionOffset : hubInfoVersionOffset+hubInfoVersionSize]),
		APIVersion: versionBigIfPresent(binary.LittleEndian.Uint32(payload[hubInfoAPIVersionOff : hubInfoAPIVersionOff+uint32Size])),
	}
	status := binary.LittleEndian.Uint32(payload[hubInfoStatusOffset : hubInfoStatusOffset+uint32Size])
	switch status {
	case sdStatusNormal:
		info.SDCardStatus = "normal"
	case sdStatusBusy:
		info.SDCardStatus = "busy"
	case sdStatusNoCard:
		info.SDCardStatus = "noCard"
	case sdStatusSlow:
		info.SDCardStatus = "slow"
	case sdStatusError:
		info.SDCardStatus = "error"
	}
	// The hub reports total/free as 32-bit values in *decimal megabytes*.
	// We convert to bytes for API consumers.
	totalMB := int64(binary.LittleEndian.Uint32(payload[hubInfoTotalOffset : hubInfoTotalOffset+uint32Size]))
	freeMB := int64(binary.LittleEndian.Uint32(payload[hubInfoFreeOffset : hubInfoFreeOffset+uint32Size]))
	if totalMB > 0 {
		total := totalMB * megabyte
		free := freeMB * megabyte
		info.SDTotalBytes = &total
		info.SDFreeBytes = &free
	}
	if len(payload) >= hubInfoMCUOffset+hubInfoMCUSize {
		info.MCUVersion = cleanFixedString(payload[hubInfoMCUOffset : hubInfoMCUOffset+hubInfoMCUSize])
	}
	return info, nil
}

func versionBig(value uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
}

func versionBigIfPresent(value uint32) string {
	if value == 0 {
		return ""
	}
	return versionBig(value)
}

func versionReversedIfPresent(value []byte) string {
	if len(value) < 4 || (value[0] == 0 && value[1] == 0 && value[2] == 0 && value[3] == 0) {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", value[3], value[2], value[1], value[0])
}

func cleanFixedString(value []byte) string {
	if end := strings.IndexByte(string(value), 0); end >= 0 {
		value = value[:end]
	}
	return strings.TrimSpace(string(value))
}
