// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

// Package uacp provides encoding/decoding and automated connection handling for
// the OPC UA Connection Protocol.
//
// To establish the connection as a client, call the Dial() function.
//
// To wait for the client to connect to, call Listen() method, and to establish
// connection with the Accept() method.
//
// Once you have a connection you can call Read() to receive full UACP messages
// including the header.
//
// In uacp, *Conn also implements Local/RemoteEndpoint() methods which returns
// EndpointURL of client or server.
//
// The data on top of UACP connection is passed as it is as long as the
// connection is established. In other words, uacp never cares the data even if
// it seems invalid. Users of this package should check the data to make sure it
// is what they want or not.
package uacp
