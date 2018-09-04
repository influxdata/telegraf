﻿/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 * Contains some contributions under the Thrift Software License.
 * Please see doc/old-thrift-license.txt in the Thrift distribution for
 * details.
 */

using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Sockets;
using System.Reflection;
using System.Text;
#if NET45
using System.Threading.Tasks;
#endif

namespace Thrift.Transport
{
    /// <summary>
    /// PropertyInfo for the DualMode property of the System.Net.Sockets.Socket class. Used to determine if the sockets are capable of
    /// automatic IPv4 and IPv6 handling. If DualMode is present the sockets automatically handle IPv4 and IPv6 connections.
    /// If the DualMode is not available the system configuration determines whether IPv4 or IPv6 is used.
    /// </summary>
    internal static class TSocketVersionizer
    {
        /// <summary>
        /// Creates a TcpClient according to the capabilitites of the used framework
        /// </summary>
        internal static TcpClient CreateTcpClient()
        {
            TcpClient client = null;

#if NET45
            client = new TcpClient(AddressFamily.InterNetworkV6);
            client.Client.DualMode = true;
#else
            client = new TcpClient(AddressFamily.InterNetwork);
#endif

            return client;
        }

        /// <summary>
        /// Creates a TcpListener according to the capabilitites of the used framework.
        /// </summary>
        internal static TcpListener CreateTcpListener(Int32 port)
        {
            TcpListener listener = null;

#if NET45
            listener = new TcpListener(System.Net.IPAddress.IPv6Any, port);
            listener.Server.DualMode = true;
#else

            listener = new TcpListener(System.Net.IPAddress.Any, port);
#endif

            return listener;
        }
    }
}
