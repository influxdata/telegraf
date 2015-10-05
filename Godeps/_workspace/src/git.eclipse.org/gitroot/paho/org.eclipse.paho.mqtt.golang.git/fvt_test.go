/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

// Use setup_IMA.sh for IBM's MessageSight
// Use fvt/rsmb.cfg for IBM's Really Small Message Broker
// Use fvt/mosquitto.cfg for the open source Mosquitto project

// Set these values to the URI of your MQTT Broker before running go-test
const (
	FVTAddr = "iot.eclipse.org"
	FVTTCP  = "tcp://" + FVTAddr + ":1883"
	FVTSSL  = "ssl://" + FVTAddr + ":8883"
)
