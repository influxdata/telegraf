/*
Copyright (c) 2014-2018 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package methods

import (
        "context"
        vsantypes "github.com/influxdata/telegraf/plugins/inputs/vsan/vsan-sdk/types"
        "github.com/vmware/govmomi/vim25/soap"
)
  type AbdicateDomOwnershipBody struct{
    Req *vsantypes.AbdicateDomOwnership `xml:"urn:vsan AbdicateDomOwnership,omitempty"`
    Res *vsantypes.AbdicateDomOwnershipResponse `xml:"urn:vsan AbdicateDomOwnershipResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AbdicateDomOwnershipBody) Fault() *soap.Fault { return b.Fault_ }

func AbdicateDomOwnership(ctx context.Context, r soap.RoundTripper, req *vsantypes.AbdicateDomOwnership) (*vsantypes.AbdicateDomOwnershipResponse, error) {
  var reqBody, resBody AbdicateDomOwnershipBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcknowledgeAlarmBody struct{
    Req *vsantypes.AcknowledgeAlarm `xml:"urn:vsan AcknowledgeAlarm,omitempty"`
    Res *vsantypes.AcknowledgeAlarmResponse `xml:"urn:vsan AcknowledgeAlarmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcknowledgeAlarmBody) Fault() *soap.Fault { return b.Fault_ }

func AcknowledgeAlarm(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcknowledgeAlarm) (*vsantypes.AcknowledgeAlarmResponse, error) {
  var reqBody, resBody AcknowledgeAlarmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcquireCimServicesTicketBody struct{
    Req *vsantypes.AcquireCimServicesTicket `xml:"urn:vsan AcquireCimServicesTicket,omitempty"`
    Res *vsantypes.AcquireCimServicesTicketResponse `xml:"urn:vsan AcquireCimServicesTicketResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcquireCimServicesTicketBody) Fault() *soap.Fault { return b.Fault_ }

func AcquireCimServicesTicket(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcquireCimServicesTicket) (*vsantypes.AcquireCimServicesTicketResponse, error) {
  var reqBody, resBody AcquireCimServicesTicketBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcquireCloneTicketBody struct{
    Req *vsantypes.AcquireCloneTicket `xml:"urn:vsan AcquireCloneTicket,omitempty"`
    Res *vsantypes.AcquireCloneTicketResponse `xml:"urn:vsan AcquireCloneTicketResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcquireCloneTicketBody) Fault() *soap.Fault { return b.Fault_ }

func AcquireCloneTicket(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcquireCloneTicket) (*vsantypes.AcquireCloneTicketResponse, error) {
  var reqBody, resBody AcquireCloneTicketBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcquireCredentialsInGuestBody struct{
    Req *vsantypes.AcquireCredentialsInGuest `xml:"urn:vsan AcquireCredentialsInGuest,omitempty"`
    Res *vsantypes.AcquireCredentialsInGuestResponse `xml:"urn:vsan AcquireCredentialsInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcquireCredentialsInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func AcquireCredentialsInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcquireCredentialsInGuest) (*vsantypes.AcquireCredentialsInGuestResponse, error) {
  var reqBody, resBody AcquireCredentialsInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcquireGenericServiceTicketBody struct{
    Req *vsantypes.AcquireGenericServiceTicket `xml:"urn:vsan AcquireGenericServiceTicket,omitempty"`
    Res *vsantypes.AcquireGenericServiceTicketResponse `xml:"urn:vsan AcquireGenericServiceTicketResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcquireGenericServiceTicketBody) Fault() *soap.Fault { return b.Fault_ }

func AcquireGenericServiceTicket(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcquireGenericServiceTicket) (*vsantypes.AcquireGenericServiceTicketResponse, error) {
  var reqBody, resBody AcquireGenericServiceTicketBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcquireLocalTicketBody struct{
    Req *vsantypes.AcquireLocalTicket `xml:"urn:vsan AcquireLocalTicket,omitempty"`
    Res *vsantypes.AcquireLocalTicketResponse `xml:"urn:vsan AcquireLocalTicketResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcquireLocalTicketBody) Fault() *soap.Fault { return b.Fault_ }

func AcquireLocalTicket(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcquireLocalTicket) (*vsantypes.AcquireLocalTicketResponse, error) {
  var reqBody, resBody AcquireLocalTicketBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcquireMksTicketBody struct{
    Req *vsantypes.AcquireMksTicket `xml:"urn:vsan AcquireMksTicket,omitempty"`
    Res *vsantypes.AcquireMksTicketResponse `xml:"urn:vsan AcquireMksTicketResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcquireMksTicketBody) Fault() *soap.Fault { return b.Fault_ }

func AcquireMksTicket(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcquireMksTicket) (*vsantypes.AcquireMksTicketResponse, error) {
  var reqBody, resBody AcquireMksTicketBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AcquireTicketBody struct{
    Req *vsantypes.AcquireTicket `xml:"urn:vsan AcquireTicket,omitempty"`
    Res *vsantypes.AcquireTicketResponse `xml:"urn:vsan AcquireTicketResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AcquireTicketBody) Fault() *soap.Fault { return b.Fault_ }

func AcquireTicket(ctx context.Context, r soap.RoundTripper, req *vsantypes.AcquireTicket) (*vsantypes.AcquireTicketResponse, error) {
  var reqBody, resBody AcquireTicketBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddAuthorizationRoleBody struct{
    Req *vsantypes.AddAuthorizationRole `xml:"urn:vsan AddAuthorizationRole,omitempty"`
    Res *vsantypes.AddAuthorizationRoleResponse `xml:"urn:vsan AddAuthorizationRoleResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddAuthorizationRoleBody) Fault() *soap.Fault { return b.Fault_ }

func AddAuthorizationRole(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddAuthorizationRole) (*vsantypes.AddAuthorizationRoleResponse, error) {
  var reqBody, resBody AddAuthorizationRoleBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddCustomFieldDefBody struct{
    Req *vsantypes.AddCustomFieldDef `xml:"urn:vsan AddCustomFieldDef,omitempty"`
    Res *vsantypes.AddCustomFieldDefResponse `xml:"urn:vsan AddCustomFieldDefResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddCustomFieldDefBody) Fault() *soap.Fault { return b.Fault_ }

func AddCustomFieldDef(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddCustomFieldDef) (*vsantypes.AddCustomFieldDefResponse, error) {
  var reqBody, resBody AddCustomFieldDefBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddDVPortgroup_TaskBody struct{
    Req *vsantypes.AddDVPortgroup_Task `xml:"urn:vsan AddDVPortgroup_Task,omitempty"`
    Res *vsantypes.AddDVPortgroup_TaskResponse `xml:"urn:vsan AddDVPortgroup_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddDVPortgroup_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func AddDVPortgroup_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddDVPortgroup_Task) (*vsantypes.AddDVPortgroup_TaskResponse, error) {
  var reqBody, resBody AddDVPortgroup_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddDisks_TaskBody struct{
    Req *vsantypes.AddDisks_Task `xml:"urn:vsan AddDisks_Task,omitempty"`
    Res *vsantypes.AddDisks_TaskResponse `xml:"urn:vsan AddDisks_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddDisks_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func AddDisks_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddDisks_Task) (*vsantypes.AddDisks_TaskResponse, error) {
  var reqBody, resBody AddDisks_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddFilterBody struct{
    Req *vsantypes.AddFilter `xml:"urn:vsan AddFilter,omitempty"`
    Res *vsantypes.AddFilterResponse `xml:"urn:vsan AddFilterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddFilterBody) Fault() *soap.Fault { return b.Fault_ }

func AddFilter(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddFilter) (*vsantypes.AddFilterResponse, error) {
  var reqBody, resBody AddFilterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddFilterEntitiesBody struct{
    Req *vsantypes.AddFilterEntities `xml:"urn:vsan AddFilterEntities,omitempty"`
    Res *vsantypes.AddFilterEntitiesResponse `xml:"urn:vsan AddFilterEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddFilterEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func AddFilterEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddFilterEntities) (*vsantypes.AddFilterEntitiesResponse, error) {
  var reqBody, resBody AddFilterEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddGuestAliasBody struct{
    Req *vsantypes.AddGuestAlias `xml:"urn:vsan AddGuestAlias,omitempty"`
    Res *vsantypes.AddGuestAliasResponse `xml:"urn:vsan AddGuestAliasResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddGuestAliasBody) Fault() *soap.Fault { return b.Fault_ }

func AddGuestAlias(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddGuestAlias) (*vsantypes.AddGuestAliasResponse, error) {
  var reqBody, resBody AddGuestAliasBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddHost_TaskBody struct{
    Req *vsantypes.AddHost_Task `xml:"urn:vsan AddHost_Task,omitempty"`
    Res *vsantypes.AddHost_TaskResponse `xml:"urn:vsan AddHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func AddHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddHost_Task) (*vsantypes.AddHost_TaskResponse, error) {
  var reqBody, resBody AddHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddInternetScsiSendTargetsBody struct{
    Req *vsantypes.AddInternetScsiSendTargets `xml:"urn:vsan AddInternetScsiSendTargets,omitempty"`
    Res *vsantypes.AddInternetScsiSendTargetsResponse `xml:"urn:vsan AddInternetScsiSendTargetsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddInternetScsiSendTargetsBody) Fault() *soap.Fault { return b.Fault_ }

func AddInternetScsiSendTargets(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddInternetScsiSendTargets) (*vsantypes.AddInternetScsiSendTargetsResponse, error) {
  var reqBody, resBody AddInternetScsiSendTargetsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddInternetScsiStaticTargetsBody struct{
    Req *vsantypes.AddInternetScsiStaticTargets `xml:"urn:vsan AddInternetScsiStaticTargets,omitempty"`
    Res *vsantypes.AddInternetScsiStaticTargetsResponse `xml:"urn:vsan AddInternetScsiStaticTargetsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddInternetScsiStaticTargetsBody) Fault() *soap.Fault { return b.Fault_ }

func AddInternetScsiStaticTargets(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddInternetScsiStaticTargets) (*vsantypes.AddInternetScsiStaticTargetsResponse, error) {
  var reqBody, resBody AddInternetScsiStaticTargetsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddKeyBody struct{
    Req *vsantypes.AddKey `xml:"urn:vsan AddKey,omitempty"`
    Res *vsantypes.AddKeyResponse `xml:"urn:vsan AddKeyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddKeyBody) Fault() *soap.Fault { return b.Fault_ }

func AddKey(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddKey) (*vsantypes.AddKeyResponse, error) {
  var reqBody, resBody AddKeyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddKeysBody struct{
    Req *vsantypes.AddKeys `xml:"urn:vsan AddKeys,omitempty"`
    Res *vsantypes.AddKeysResponse `xml:"urn:vsan AddKeysResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddKeysBody) Fault() *soap.Fault { return b.Fault_ }

func AddKeys(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddKeys) (*vsantypes.AddKeysResponse, error) {
  var reqBody, resBody AddKeysBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddLicenseBody struct{
    Req *vsantypes.AddLicense `xml:"urn:vsan AddLicense,omitempty"`
    Res *vsantypes.AddLicenseResponse `xml:"urn:vsan AddLicenseResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddLicenseBody) Fault() *soap.Fault { return b.Fault_ }

func AddLicense(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddLicense) (*vsantypes.AddLicenseResponse, error) {
  var reqBody, resBody AddLicenseBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddMonitoredEntitiesBody struct{
    Req *vsantypes.AddMonitoredEntities `xml:"urn:vsan AddMonitoredEntities,omitempty"`
    Res *vsantypes.AddMonitoredEntitiesResponse `xml:"urn:vsan AddMonitoredEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddMonitoredEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func AddMonitoredEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddMonitoredEntities) (*vsantypes.AddMonitoredEntitiesResponse, error) {
  var reqBody, resBody AddMonitoredEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddNetworkResourcePoolBody struct{
    Req *vsantypes.AddNetworkResourcePool `xml:"urn:vsan AddNetworkResourcePool,omitempty"`
    Res *vsantypes.AddNetworkResourcePoolResponse `xml:"urn:vsan AddNetworkResourcePoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddNetworkResourcePoolBody) Fault() *soap.Fault { return b.Fault_ }

func AddNetworkResourcePool(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddNetworkResourcePool) (*vsantypes.AddNetworkResourcePoolResponse, error) {
  var reqBody, resBody AddNetworkResourcePoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddPortGroupBody struct{
    Req *vsantypes.AddPortGroup `xml:"urn:vsan AddPortGroup,omitempty"`
    Res *vsantypes.AddPortGroupResponse `xml:"urn:vsan AddPortGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddPortGroupBody) Fault() *soap.Fault { return b.Fault_ }

func AddPortGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddPortGroup) (*vsantypes.AddPortGroupResponse, error) {
  var reqBody, resBody AddPortGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddServiceConsoleVirtualNicBody struct{
    Req *vsantypes.AddServiceConsoleVirtualNic `xml:"urn:vsan AddServiceConsoleVirtualNic,omitempty"`
    Res *vsantypes.AddServiceConsoleVirtualNicResponse `xml:"urn:vsan AddServiceConsoleVirtualNicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddServiceConsoleVirtualNicBody) Fault() *soap.Fault { return b.Fault_ }

func AddServiceConsoleVirtualNic(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddServiceConsoleVirtualNic) (*vsantypes.AddServiceConsoleVirtualNicResponse, error) {
  var reqBody, resBody AddServiceConsoleVirtualNicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddStandaloneHost_TaskBody struct{
    Req *vsantypes.AddStandaloneHost_Task `xml:"urn:vsan AddStandaloneHost_Task,omitempty"`
    Res *vsantypes.AddStandaloneHost_TaskResponse `xml:"urn:vsan AddStandaloneHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddStandaloneHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func AddStandaloneHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddStandaloneHost_Task) (*vsantypes.AddStandaloneHost_TaskResponse, error) {
  var reqBody, resBody AddStandaloneHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddVirtualNicBody struct{
    Req *vsantypes.AddVirtualNic `xml:"urn:vsan AddVirtualNic,omitempty"`
    Res *vsantypes.AddVirtualNicResponse `xml:"urn:vsan AddVirtualNicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddVirtualNicBody) Fault() *soap.Fault { return b.Fault_ }

func AddVirtualNic(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddVirtualNic) (*vsantypes.AddVirtualNicResponse, error) {
  var reqBody, resBody AddVirtualNicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AddVirtualSwitchBody struct{
    Req *vsantypes.AddVirtualSwitch `xml:"urn:vsan AddVirtualSwitch,omitempty"`
    Res *vsantypes.AddVirtualSwitchResponse `xml:"urn:vsan AddVirtualSwitchResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AddVirtualSwitchBody) Fault() *soap.Fault { return b.Fault_ }

func AddVirtualSwitch(ctx context.Context, r soap.RoundTripper, req *vsantypes.AddVirtualSwitch) (*vsantypes.AddVirtualSwitchResponse, error) {
  var reqBody, resBody AddVirtualSwitchBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AllocateIpv4AddressBody struct{
    Req *vsantypes.AllocateIpv4Address `xml:"urn:vsan AllocateIpv4Address,omitempty"`
    Res *vsantypes.AllocateIpv4AddressResponse `xml:"urn:vsan AllocateIpv4AddressResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AllocateIpv4AddressBody) Fault() *soap.Fault { return b.Fault_ }

func AllocateIpv4Address(ctx context.Context, r soap.RoundTripper, req *vsantypes.AllocateIpv4Address) (*vsantypes.AllocateIpv4AddressResponse, error) {
  var reqBody, resBody AllocateIpv4AddressBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AllocateIpv6AddressBody struct{
    Req *vsantypes.AllocateIpv6Address `xml:"urn:vsan AllocateIpv6Address,omitempty"`
    Res *vsantypes.AllocateIpv6AddressResponse `xml:"urn:vsan AllocateIpv6AddressResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AllocateIpv6AddressBody) Fault() *soap.Fault { return b.Fault_ }

func AllocateIpv6Address(ctx context.Context, r soap.RoundTripper, req *vsantypes.AllocateIpv6Address) (*vsantypes.AllocateIpv6AddressResponse, error) {
  var reqBody, resBody AllocateIpv6AddressBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AnswerVMBody struct{
    Req *vsantypes.AnswerVM `xml:"urn:vsan AnswerVM,omitempty"`
    Res *vsantypes.AnswerVMResponse `xml:"urn:vsan AnswerVMResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AnswerVMBody) Fault() *soap.Fault { return b.Fault_ }

func AnswerVM(ctx context.Context, r soap.RoundTripper, req *vsantypes.AnswerVM) (*vsantypes.AnswerVMResponse, error) {
  var reqBody, resBody AnswerVMBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ApplyEntitiesConfig_TaskBody struct{
    Req *vsantypes.ApplyEntitiesConfig_Task `xml:"urn:vsan ApplyEntitiesConfig_Task,omitempty"`
    Res *vsantypes.ApplyEntitiesConfig_TaskResponse `xml:"urn:vsan ApplyEntitiesConfig_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ApplyEntitiesConfig_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ApplyEntitiesConfig_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ApplyEntitiesConfig_Task) (*vsantypes.ApplyEntitiesConfig_TaskResponse, error) {
  var reqBody, resBody ApplyEntitiesConfig_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ApplyHostConfig_TaskBody struct{
    Req *vsantypes.ApplyHostConfig_Task `xml:"urn:vsan ApplyHostConfig_Task,omitempty"`
    Res *vsantypes.ApplyHostConfig_TaskResponse `xml:"urn:vsan ApplyHostConfig_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ApplyHostConfig_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ApplyHostConfig_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ApplyHostConfig_Task) (*vsantypes.ApplyHostConfig_TaskResponse, error) {
  var reqBody, resBody ApplyHostConfig_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ApplyRecommendationBody struct{
    Req *vsantypes.ApplyRecommendation `xml:"urn:vsan ApplyRecommendation,omitempty"`
    Res *vsantypes.ApplyRecommendationResponse `xml:"urn:vsan ApplyRecommendationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ApplyRecommendationBody) Fault() *soap.Fault { return b.Fault_ }

func ApplyRecommendation(ctx context.Context, r soap.RoundTripper, req *vsantypes.ApplyRecommendation) (*vsantypes.ApplyRecommendationResponse, error) {
  var reqBody, resBody ApplyRecommendationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ApplyStorageDrsRecommendationToPod_TaskBody struct{
    Req *vsantypes.ApplyStorageDrsRecommendationToPod_Task `xml:"urn:vsan ApplyStorageDrsRecommendationToPod_Task,omitempty"`
    Res *vsantypes.ApplyStorageDrsRecommendationToPod_TaskResponse `xml:"urn:vsan ApplyStorageDrsRecommendationToPod_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ApplyStorageDrsRecommendationToPod_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ApplyStorageDrsRecommendationToPod_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ApplyStorageDrsRecommendationToPod_Task) (*vsantypes.ApplyStorageDrsRecommendationToPod_TaskResponse, error) {
  var reqBody, resBody ApplyStorageDrsRecommendationToPod_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ApplyStorageDrsRecommendation_TaskBody struct{
    Req *vsantypes.ApplyStorageDrsRecommendation_Task `xml:"urn:vsan ApplyStorageDrsRecommendation_Task,omitempty"`
    Res *vsantypes.ApplyStorageDrsRecommendation_TaskResponse `xml:"urn:vsan ApplyStorageDrsRecommendation_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ApplyStorageDrsRecommendation_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ApplyStorageDrsRecommendation_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ApplyStorageDrsRecommendation_Task) (*vsantypes.ApplyStorageDrsRecommendation_TaskResponse, error) {
  var reqBody, resBody ApplyStorageDrsRecommendation_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AreAlarmActionsEnabledBody struct{
    Req *vsantypes.AreAlarmActionsEnabled `xml:"urn:vsan AreAlarmActionsEnabled,omitempty"`
    Res *vsantypes.AreAlarmActionsEnabledResponse `xml:"urn:vsan AreAlarmActionsEnabledResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AreAlarmActionsEnabledBody) Fault() *soap.Fault { return b.Fault_ }

func AreAlarmActionsEnabled(ctx context.Context, r soap.RoundTripper, req *vsantypes.AreAlarmActionsEnabled) (*vsantypes.AreAlarmActionsEnabledResponse, error) {
  var reqBody, resBody AreAlarmActionsEnabledBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AssignUserToGroupBody struct{
    Req *vsantypes.AssignUserToGroup `xml:"urn:vsan AssignUserToGroup,omitempty"`
    Res *vsantypes.AssignUserToGroupResponse `xml:"urn:vsan AssignUserToGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AssignUserToGroupBody) Fault() *soap.Fault { return b.Fault_ }

func AssignUserToGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.AssignUserToGroup) (*vsantypes.AssignUserToGroupResponse, error) {
  var reqBody, resBody AssignUserToGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AssociateProfileBody struct{
    Req *vsantypes.AssociateProfile `xml:"urn:vsan AssociateProfile,omitempty"`
    Res *vsantypes.AssociateProfileResponse `xml:"urn:vsan AssociateProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AssociateProfileBody) Fault() *soap.Fault { return b.Fault_ }

func AssociateProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.AssociateProfile) (*vsantypes.AssociateProfileResponse, error) {
  var reqBody, resBody AssociateProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AttachDisk_TaskBody struct{
    Req *vsantypes.AttachDisk_Task `xml:"urn:vsan AttachDisk_Task,omitempty"`
    Res *vsantypes.AttachDisk_TaskResponse `xml:"urn:vsan AttachDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AttachDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func AttachDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.AttachDisk_Task) (*vsantypes.AttachDisk_TaskResponse, error) {
  var reqBody, resBody AttachDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AttachScsiLunBody struct{
    Req *vsantypes.AttachScsiLun `xml:"urn:vsan AttachScsiLun,omitempty"`
    Res *vsantypes.AttachScsiLunResponse `xml:"urn:vsan AttachScsiLunResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AttachScsiLunBody) Fault() *soap.Fault { return b.Fault_ }

func AttachScsiLun(ctx context.Context, r soap.RoundTripper, req *vsantypes.AttachScsiLun) (*vsantypes.AttachScsiLunResponse, error) {
  var reqBody, resBody AttachScsiLunBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AttachScsiLunEx_TaskBody struct{
    Req *vsantypes.AttachScsiLunEx_Task `xml:"urn:vsan AttachScsiLunEx_Task,omitempty"`
    Res *vsantypes.AttachScsiLunEx_TaskResponse `xml:"urn:vsan AttachScsiLunEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AttachScsiLunEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func AttachScsiLunEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.AttachScsiLunEx_Task) (*vsantypes.AttachScsiLunEx_TaskResponse, error) {
  var reqBody, resBody AttachScsiLunEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AttachTagToVStorageObjectBody struct{
    Req *vsantypes.AttachTagToVStorageObject `xml:"urn:vsan AttachTagToVStorageObject,omitempty"`
    Res *vsantypes.AttachTagToVStorageObjectResponse `xml:"urn:vsan AttachTagToVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AttachTagToVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func AttachTagToVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.AttachTagToVStorageObject) (*vsantypes.AttachTagToVStorageObjectResponse, error) {
  var reqBody, resBody AttachTagToVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AttachVmfsExtentBody struct{
    Req *vsantypes.AttachVmfsExtent `xml:"urn:vsan AttachVmfsExtent,omitempty"`
    Res *vsantypes.AttachVmfsExtentResponse `xml:"urn:vsan AttachVmfsExtentResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AttachVmfsExtentBody) Fault() *soap.Fault { return b.Fault_ }

func AttachVmfsExtent(ctx context.Context, r soap.RoundTripper, req *vsantypes.AttachVmfsExtent) (*vsantypes.AttachVmfsExtentResponse, error) {
  var reqBody, resBody AttachVmfsExtentBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AutoStartPowerOffBody struct{
    Req *vsantypes.AutoStartPowerOff `xml:"urn:vsan AutoStartPowerOff,omitempty"`
    Res *vsantypes.AutoStartPowerOffResponse `xml:"urn:vsan AutoStartPowerOffResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AutoStartPowerOffBody) Fault() *soap.Fault { return b.Fault_ }

func AutoStartPowerOff(ctx context.Context, r soap.RoundTripper, req *vsantypes.AutoStartPowerOff) (*vsantypes.AutoStartPowerOffResponse, error) {
  var reqBody, resBody AutoStartPowerOffBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type AutoStartPowerOnBody struct{
    Req *vsantypes.AutoStartPowerOn `xml:"urn:vsan AutoStartPowerOn,omitempty"`
    Res *vsantypes.AutoStartPowerOnResponse `xml:"urn:vsan AutoStartPowerOnResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *AutoStartPowerOnBody) Fault() *soap.Fault { return b.Fault_ }

func AutoStartPowerOn(ctx context.Context, r soap.RoundTripper, req *vsantypes.AutoStartPowerOn) (*vsantypes.AutoStartPowerOnResponse, error) {
  var reqBody, resBody AutoStartPowerOnBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type BackupFirmwareConfigurationBody struct{
    Req *vsantypes.BackupFirmwareConfiguration `xml:"urn:vsan BackupFirmwareConfiguration,omitempty"`
    Res *vsantypes.BackupFirmwareConfigurationResponse `xml:"urn:vsan BackupFirmwareConfigurationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *BackupFirmwareConfigurationBody) Fault() *soap.Fault { return b.Fault_ }

func BackupFirmwareConfiguration(ctx context.Context, r soap.RoundTripper, req *vsantypes.BackupFirmwareConfiguration) (*vsantypes.BackupFirmwareConfigurationResponse, error) {
  var reqBody, resBody BackupFirmwareConfigurationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type BindVnicBody struct{
    Req *vsantypes.BindVnic `xml:"urn:vsan BindVnic,omitempty"`
    Res *vsantypes.BindVnicResponse `xml:"urn:vsan BindVnicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *BindVnicBody) Fault() *soap.Fault { return b.Fault_ }

func BindVnic(ctx context.Context, r soap.RoundTripper, req *vsantypes.BindVnic) (*vsantypes.BindVnicResponse, error) {
  var reqBody, resBody BindVnicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type BrowseDiagnosticLogBody struct{
    Req *vsantypes.BrowseDiagnosticLog `xml:"urn:vsan BrowseDiagnosticLog,omitempty"`
    Res *vsantypes.BrowseDiagnosticLogResponse `xml:"urn:vsan BrowseDiagnosticLogResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *BrowseDiagnosticLogBody) Fault() *soap.Fault { return b.Fault_ }

func BrowseDiagnosticLog(ctx context.Context, r soap.RoundTripper, req *vsantypes.BrowseDiagnosticLog) (*vsantypes.BrowseDiagnosticLogResponse, error) {
  var reqBody, resBody BrowseDiagnosticLogBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CanProvisionObjectsBody struct{
    Req *vsantypes.CanProvisionObjects `xml:"urn:vsan CanProvisionObjects,omitempty"`
    Res *vsantypes.CanProvisionObjectsResponse `xml:"urn:vsan CanProvisionObjectsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CanProvisionObjectsBody) Fault() *soap.Fault { return b.Fault_ }

func CanProvisionObjects(ctx context.Context, r soap.RoundTripper, req *vsantypes.CanProvisionObjects) (*vsantypes.CanProvisionObjectsResponse, error) {
  var reqBody, resBody CanProvisionObjectsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CancelRecommendationBody struct{
    Req *vsantypes.CancelRecommendation `xml:"urn:vsan CancelRecommendation,omitempty"`
    Res *vsantypes.CancelRecommendationResponse `xml:"urn:vsan CancelRecommendationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CancelRecommendationBody) Fault() *soap.Fault { return b.Fault_ }

func CancelRecommendation(ctx context.Context, r soap.RoundTripper, req *vsantypes.CancelRecommendation) (*vsantypes.CancelRecommendationResponse, error) {
  var reqBody, resBody CancelRecommendationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CancelRetrievePropertiesExBody struct{
    Req *vsantypes.CancelRetrievePropertiesEx `xml:"urn:vsan CancelRetrievePropertiesEx,omitempty"`
    Res *vsantypes.CancelRetrievePropertiesExResponse `xml:"urn:vsan CancelRetrievePropertiesExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CancelRetrievePropertiesExBody) Fault() *soap.Fault { return b.Fault_ }

func CancelRetrievePropertiesEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.CancelRetrievePropertiesEx) (*vsantypes.CancelRetrievePropertiesExResponse, error) {
  var reqBody, resBody CancelRetrievePropertiesExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CancelStorageDrsRecommendationBody struct{
    Req *vsantypes.CancelStorageDrsRecommendation `xml:"urn:vsan CancelStorageDrsRecommendation,omitempty"`
    Res *vsantypes.CancelStorageDrsRecommendationResponse `xml:"urn:vsan CancelStorageDrsRecommendationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CancelStorageDrsRecommendationBody) Fault() *soap.Fault { return b.Fault_ }

func CancelStorageDrsRecommendation(ctx context.Context, r soap.RoundTripper, req *vsantypes.CancelStorageDrsRecommendation) (*vsantypes.CancelStorageDrsRecommendationResponse, error) {
  var reqBody, resBody CancelStorageDrsRecommendationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CancelTaskBody struct{
    Req *vsantypes.CancelTask `xml:"urn:vsan CancelTask,omitempty"`
    Res *vsantypes.CancelTaskResponse `xml:"urn:vsan CancelTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CancelTaskBody) Fault() *soap.Fault { return b.Fault_ }

func CancelTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.CancelTask) (*vsantypes.CancelTaskResponse, error) {
  var reqBody, resBody CancelTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CancelWaitForUpdatesBody struct{
    Req *vsantypes.CancelWaitForUpdates `xml:"urn:vsan CancelWaitForUpdates,omitempty"`
    Res *vsantypes.CancelWaitForUpdatesResponse `xml:"urn:vsan CancelWaitForUpdatesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CancelWaitForUpdatesBody) Fault() *soap.Fault { return b.Fault_ }

func CancelWaitForUpdates(ctx context.Context, r soap.RoundTripper, req *vsantypes.CancelWaitForUpdates) (*vsantypes.CancelWaitForUpdatesResponse, error) {
  var reqBody, resBody CancelWaitForUpdatesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CertMgrRefreshCACertificatesAndCRLs_TaskBody struct{
    Req *vsantypes.CertMgrRefreshCACertificatesAndCRLs_Task `xml:"urn:vsan CertMgrRefreshCACertificatesAndCRLs_Task,omitempty"`
    Res *vsantypes.CertMgrRefreshCACertificatesAndCRLs_TaskResponse `xml:"urn:vsan CertMgrRefreshCACertificatesAndCRLs_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CertMgrRefreshCACertificatesAndCRLs_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CertMgrRefreshCACertificatesAndCRLs_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CertMgrRefreshCACertificatesAndCRLs_Task) (*vsantypes.CertMgrRefreshCACertificatesAndCRLs_TaskResponse, error) {
  var reqBody, resBody CertMgrRefreshCACertificatesAndCRLs_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CertMgrRefreshCertificates_TaskBody struct{
    Req *vsantypes.CertMgrRefreshCertificates_Task `xml:"urn:vsan CertMgrRefreshCertificates_Task,omitempty"`
    Res *vsantypes.CertMgrRefreshCertificates_TaskResponse `xml:"urn:vsan CertMgrRefreshCertificates_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CertMgrRefreshCertificates_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CertMgrRefreshCertificates_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CertMgrRefreshCertificates_Task) (*vsantypes.CertMgrRefreshCertificates_TaskResponse, error) {
  var reqBody, resBody CertMgrRefreshCertificates_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CertMgrRevokeCertificates_TaskBody struct{
    Req *vsantypes.CertMgrRevokeCertificates_Task `xml:"urn:vsan CertMgrRevokeCertificates_Task,omitempty"`
    Res *vsantypes.CertMgrRevokeCertificates_TaskResponse `xml:"urn:vsan CertMgrRevokeCertificates_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CertMgrRevokeCertificates_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CertMgrRevokeCertificates_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CertMgrRevokeCertificates_Task) (*vsantypes.CertMgrRevokeCertificates_TaskResponse, error) {
  var reqBody, resBody CertMgrRevokeCertificates_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ChangeAccessModeBody struct{
    Req *vsantypes.ChangeAccessMode `xml:"urn:vsan ChangeAccessMode,omitempty"`
    Res *vsantypes.ChangeAccessModeResponse `xml:"urn:vsan ChangeAccessModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ChangeAccessModeBody) Fault() *soap.Fault { return b.Fault_ }

func ChangeAccessMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.ChangeAccessMode) (*vsantypes.ChangeAccessModeResponse, error) {
  var reqBody, resBody ChangeAccessModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ChangeFileAttributesInGuestBody struct{
    Req *vsantypes.ChangeFileAttributesInGuest `xml:"urn:vsan ChangeFileAttributesInGuest,omitempty"`
    Res *vsantypes.ChangeFileAttributesInGuestResponse `xml:"urn:vsan ChangeFileAttributesInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ChangeFileAttributesInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ChangeFileAttributesInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ChangeFileAttributesInGuest) (*vsantypes.ChangeFileAttributesInGuestResponse, error) {
  var reqBody, resBody ChangeFileAttributesInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ChangeLockdownModeBody struct{
    Req *vsantypes.ChangeLockdownMode `xml:"urn:vsan ChangeLockdownMode,omitempty"`
    Res *vsantypes.ChangeLockdownModeResponse `xml:"urn:vsan ChangeLockdownModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ChangeLockdownModeBody) Fault() *soap.Fault { return b.Fault_ }

func ChangeLockdownMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.ChangeLockdownMode) (*vsantypes.ChangeLockdownModeResponse, error) {
  var reqBody, resBody ChangeLockdownModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ChangeNFSUserPasswordBody struct{
    Req *vsantypes.ChangeNFSUserPassword `xml:"urn:vsan ChangeNFSUserPassword,omitempty"`
    Res *vsantypes.ChangeNFSUserPasswordResponse `xml:"urn:vsan ChangeNFSUserPasswordResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ChangeNFSUserPasswordBody) Fault() *soap.Fault { return b.Fault_ }

func ChangeNFSUserPassword(ctx context.Context, r soap.RoundTripper, req *vsantypes.ChangeNFSUserPassword) (*vsantypes.ChangeNFSUserPasswordResponse, error) {
  var reqBody, resBody ChangeNFSUserPasswordBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ChangeOwnerBody struct{
    Req *vsantypes.ChangeOwner `xml:"urn:vsan ChangeOwner,omitempty"`
    Res *vsantypes.ChangeOwnerResponse `xml:"urn:vsan ChangeOwnerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ChangeOwnerBody) Fault() *soap.Fault { return b.Fault_ }

func ChangeOwner(ctx context.Context, r soap.RoundTripper, req *vsantypes.ChangeOwner) (*vsantypes.ChangeOwnerResponse, error) {
  var reqBody, resBody ChangeOwnerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckAddHostEvc_TaskBody struct{
    Req *vsantypes.CheckAddHostEvc_Task `xml:"urn:vsan CheckAddHostEvc_Task,omitempty"`
    Res *vsantypes.CheckAddHostEvc_TaskResponse `xml:"urn:vsan CheckAddHostEvc_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckAddHostEvc_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckAddHostEvc_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckAddHostEvc_Task) (*vsantypes.CheckAddHostEvc_TaskResponse, error) {
  var reqBody, resBody CheckAddHostEvc_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckAnswerFileStatus_TaskBody struct{
    Req *vsantypes.CheckAnswerFileStatus_Task `xml:"urn:vsan CheckAnswerFileStatus_Task,omitempty"`
    Res *vsantypes.CheckAnswerFileStatus_TaskResponse `xml:"urn:vsan CheckAnswerFileStatus_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckAnswerFileStatus_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckAnswerFileStatus_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckAnswerFileStatus_Task) (*vsantypes.CheckAnswerFileStatus_TaskResponse, error) {
  var reqBody, resBody CheckAnswerFileStatus_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckCompatibility_TaskBody struct{
    Req *vsantypes.CheckCompatibility_Task `xml:"urn:vsan CheckCompatibility_Task,omitempty"`
    Res *vsantypes.CheckCompatibility_TaskResponse `xml:"urn:vsan CheckCompatibility_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckCompatibility_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckCompatibility_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckCompatibility_Task) (*vsantypes.CheckCompatibility_TaskResponse, error) {
  var reqBody, resBody CheckCompatibility_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckCompliance_TaskBody struct{
    Req *vsantypes.CheckCompliance_Task `xml:"urn:vsan CheckCompliance_Task,omitempty"`
    Res *vsantypes.CheckCompliance_TaskResponse `xml:"urn:vsan CheckCompliance_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckCompliance_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckCompliance_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckCompliance_Task) (*vsantypes.CheckCompliance_TaskResponse, error) {
  var reqBody, resBody CheckCompliance_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckConfigureEvcMode_TaskBody struct{
    Req *vsantypes.CheckConfigureEvcMode_Task `xml:"urn:vsan CheckConfigureEvcMode_Task,omitempty"`
    Res *vsantypes.CheckConfigureEvcMode_TaskResponse `xml:"urn:vsan CheckConfigureEvcMode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckConfigureEvcMode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckConfigureEvcMode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckConfigureEvcMode_Task) (*vsantypes.CheckConfigureEvcMode_TaskResponse, error) {
  var reqBody, resBody CheckConfigureEvcMode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckCustomizationResourcesBody struct{
    Req *vsantypes.CheckCustomizationResources `xml:"urn:vsan CheckCustomizationResources,omitempty"`
    Res *vsantypes.CheckCustomizationResourcesResponse `xml:"urn:vsan CheckCustomizationResourcesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckCustomizationResourcesBody) Fault() *soap.Fault { return b.Fault_ }

func CheckCustomizationResources(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckCustomizationResources) (*vsantypes.CheckCustomizationResourcesResponse, error) {
  var reqBody, resBody CheckCustomizationResourcesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckCustomizationSpecBody struct{
    Req *vsantypes.CheckCustomizationSpec `xml:"urn:vsan CheckCustomizationSpec,omitempty"`
    Res *vsantypes.CheckCustomizationSpecResponse `xml:"urn:vsan CheckCustomizationSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckCustomizationSpecBody) Fault() *soap.Fault { return b.Fault_ }

func CheckCustomizationSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckCustomizationSpec) (*vsantypes.CheckCustomizationSpecResponse, error) {
  var reqBody, resBody CheckCustomizationSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckForUpdatesBody struct{
    Req *vsantypes.CheckForUpdates `xml:"urn:vsan CheckForUpdates,omitempty"`
    Res *vsantypes.CheckForUpdatesResponse `xml:"urn:vsan CheckForUpdatesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckForUpdatesBody) Fault() *soap.Fault { return b.Fault_ }

func CheckForUpdates(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckForUpdates) (*vsantypes.CheckForUpdatesResponse, error) {
  var reqBody, resBody CheckForUpdatesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckHostPatch_TaskBody struct{
    Req *vsantypes.CheckHostPatch_Task `xml:"urn:vsan CheckHostPatch_Task,omitempty"`
    Res *vsantypes.CheckHostPatch_TaskResponse `xml:"urn:vsan CheckHostPatch_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckHostPatch_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckHostPatch_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckHostPatch_Task) (*vsantypes.CheckHostPatch_TaskResponse, error) {
  var reqBody, resBody CheckHostPatch_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckLicenseFeatureBody struct{
    Req *vsantypes.CheckLicenseFeature `xml:"urn:vsan CheckLicenseFeature,omitempty"`
    Res *vsantypes.CheckLicenseFeatureResponse `xml:"urn:vsan CheckLicenseFeatureResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckLicenseFeatureBody) Fault() *soap.Fault { return b.Fault_ }

func CheckLicenseFeature(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckLicenseFeature) (*vsantypes.CheckLicenseFeatureResponse, error) {
  var reqBody, resBody CheckLicenseFeatureBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckMigrate_TaskBody struct{
    Req *vsantypes.CheckMigrate_Task `xml:"urn:vsan CheckMigrate_Task,omitempty"`
    Res *vsantypes.CheckMigrate_TaskResponse `xml:"urn:vsan CheckMigrate_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckMigrate_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckMigrate_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckMigrate_Task) (*vsantypes.CheckMigrate_TaskResponse, error) {
  var reqBody, resBody CheckMigrate_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckProfileCompliance_TaskBody struct{
    Req *vsantypes.CheckProfileCompliance_Task `xml:"urn:vsan CheckProfileCompliance_Task,omitempty"`
    Res *vsantypes.CheckProfileCompliance_TaskResponse `xml:"urn:vsan CheckProfileCompliance_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckProfileCompliance_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckProfileCompliance_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckProfileCompliance_Task) (*vsantypes.CheckProfileCompliance_TaskResponse, error) {
  var reqBody, resBody CheckProfileCompliance_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CheckRelocate_TaskBody struct{
    Req *vsantypes.CheckRelocate_Task `xml:"urn:vsan CheckRelocate_Task,omitempty"`
    Res *vsantypes.CheckRelocate_TaskResponse `xml:"urn:vsan CheckRelocate_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CheckRelocate_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CheckRelocate_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CheckRelocate_Task) (*vsantypes.CheckRelocate_TaskResponse, error) {
  var reqBody, resBody CheckRelocate_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ClearComplianceStatusBody struct{
    Req *vsantypes.ClearComplianceStatus `xml:"urn:vsan ClearComplianceStatus,omitempty"`
    Res *vsantypes.ClearComplianceStatusResponse `xml:"urn:vsan ClearComplianceStatusResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ClearComplianceStatusBody) Fault() *soap.Fault { return b.Fault_ }

func ClearComplianceStatus(ctx context.Context, r soap.RoundTripper, req *vsantypes.ClearComplianceStatus) (*vsantypes.ClearComplianceStatusResponse, error) {
  var reqBody, resBody ClearComplianceStatusBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ClearNFSUserBody struct{
    Req *vsantypes.ClearNFSUser `xml:"urn:vsan ClearNFSUser,omitempty"`
    Res *vsantypes.ClearNFSUserResponse `xml:"urn:vsan ClearNFSUserResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ClearNFSUserBody) Fault() *soap.Fault { return b.Fault_ }

func ClearNFSUser(ctx context.Context, r soap.RoundTripper, req *vsantypes.ClearNFSUser) (*vsantypes.ClearNFSUserResponse, error) {
  var reqBody, resBody ClearNFSUserBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ClearSystemEventLogBody struct{
    Req *vsantypes.ClearSystemEventLog `xml:"urn:vsan ClearSystemEventLog,omitempty"`
    Res *vsantypes.ClearSystemEventLogResponse `xml:"urn:vsan ClearSystemEventLogResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ClearSystemEventLogBody) Fault() *soap.Fault { return b.Fault_ }

func ClearSystemEventLog(ctx context.Context, r soap.RoundTripper, req *vsantypes.ClearSystemEventLog) (*vsantypes.ClearSystemEventLogResponse, error) {
  var reqBody, resBody ClearSystemEventLogBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CloneSessionBody struct{
    Req *vsantypes.CloneSession `xml:"urn:vsan CloneSession,omitempty"`
    Res *vsantypes.CloneSessionResponse `xml:"urn:vsan CloneSessionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CloneSessionBody) Fault() *soap.Fault { return b.Fault_ }

func CloneSession(ctx context.Context, r soap.RoundTripper, req *vsantypes.CloneSession) (*vsantypes.CloneSessionResponse, error) {
  var reqBody, resBody CloneSessionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CloneVApp_TaskBody struct{
    Req *vsantypes.CloneVApp_Task `xml:"urn:vsan CloneVApp_Task,omitempty"`
    Res *vsantypes.CloneVApp_TaskResponse `xml:"urn:vsan CloneVApp_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CloneVApp_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CloneVApp_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CloneVApp_Task) (*vsantypes.CloneVApp_TaskResponse, error) {
  var reqBody, resBody CloneVApp_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CloneVM_TaskBody struct{
    Req *vsantypes.CloneVM_Task `xml:"urn:vsan CloneVM_Task,omitempty"`
    Res *vsantypes.CloneVM_TaskResponse `xml:"urn:vsan CloneVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CloneVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CloneVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CloneVM_Task) (*vsantypes.CloneVM_TaskResponse, error) {
  var reqBody, resBody CloneVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CloneVStorageObject_TaskBody struct{
    Req *vsantypes.CloneVStorageObject_Task `xml:"urn:vsan CloneVStorageObject_Task,omitempty"`
    Res *vsantypes.CloneVStorageObject_TaskResponse `xml:"urn:vsan CloneVStorageObject_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CloneVStorageObject_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CloneVStorageObject_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CloneVStorageObject_Task) (*vsantypes.CloneVStorageObject_TaskResponse, error) {
  var reqBody, resBody CloneVStorageObject_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CloseInventoryViewFolderBody struct{
    Req *vsantypes.CloseInventoryViewFolder `xml:"urn:vsan CloseInventoryViewFolder,omitempty"`
    Res *vsantypes.CloseInventoryViewFolderResponse `xml:"urn:vsan CloseInventoryViewFolderResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CloseInventoryViewFolderBody) Fault() *soap.Fault { return b.Fault_ }

func CloseInventoryViewFolder(ctx context.Context, r soap.RoundTripper, req *vsantypes.CloseInventoryViewFolder) (*vsantypes.CloseInventoryViewFolderResponse, error) {
  var reqBody, resBody CloseInventoryViewFolderBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ClusterEnterMaintenanceModeBody struct{
    Req *vsantypes.ClusterEnterMaintenanceMode `xml:"urn:vsan ClusterEnterMaintenanceMode,omitempty"`
    Res *vsantypes.ClusterEnterMaintenanceModeResponse `xml:"urn:vsan ClusterEnterMaintenanceModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ClusterEnterMaintenanceModeBody) Fault() *soap.Fault { return b.Fault_ }

func ClusterEnterMaintenanceMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.ClusterEnterMaintenanceMode) (*vsantypes.ClusterEnterMaintenanceModeResponse, error) {
  var reqBody, resBody ClusterEnterMaintenanceModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ComputeDiskPartitionInfoBody struct{
    Req *vsantypes.ComputeDiskPartitionInfo `xml:"urn:vsan ComputeDiskPartitionInfo,omitempty"`
    Res *vsantypes.ComputeDiskPartitionInfoResponse `xml:"urn:vsan ComputeDiskPartitionInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ComputeDiskPartitionInfoBody) Fault() *soap.Fault { return b.Fault_ }

func ComputeDiskPartitionInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.ComputeDiskPartitionInfo) (*vsantypes.ComputeDiskPartitionInfoResponse, error) {
  var reqBody, resBody ComputeDiskPartitionInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ComputeDiskPartitionInfoForResizeBody struct{
    Req *vsantypes.ComputeDiskPartitionInfoForResize `xml:"urn:vsan ComputeDiskPartitionInfoForResize,omitempty"`
    Res *vsantypes.ComputeDiskPartitionInfoForResizeResponse `xml:"urn:vsan ComputeDiskPartitionInfoForResizeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ComputeDiskPartitionInfoForResizeBody) Fault() *soap.Fault { return b.Fault_ }

func ComputeDiskPartitionInfoForResize(ctx context.Context, r soap.RoundTripper, req *vsantypes.ComputeDiskPartitionInfoForResize) (*vsantypes.ComputeDiskPartitionInfoForResizeResponse, error) {
  var reqBody, resBody ComputeDiskPartitionInfoForResizeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureCryptoKeyBody struct{
    Req *vsantypes.ConfigureCryptoKey `xml:"urn:vsan ConfigureCryptoKey,omitempty"`
    Res *vsantypes.ConfigureCryptoKeyResponse `xml:"urn:vsan ConfigureCryptoKeyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureCryptoKeyBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureCryptoKey(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureCryptoKey) (*vsantypes.ConfigureCryptoKeyResponse, error) {
  var reqBody, resBody ConfigureCryptoKeyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureDatastoreIORM_TaskBody struct{
    Req *vsantypes.ConfigureDatastoreIORM_Task `xml:"urn:vsan ConfigureDatastoreIORM_Task,omitempty"`
    Res *vsantypes.ConfigureDatastoreIORM_TaskResponse `xml:"urn:vsan ConfigureDatastoreIORM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureDatastoreIORM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureDatastoreIORM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureDatastoreIORM_Task) (*vsantypes.ConfigureDatastoreIORM_TaskResponse, error) {
  var reqBody, resBody ConfigureDatastoreIORM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureDatastorePrincipalBody struct{
    Req *vsantypes.ConfigureDatastorePrincipal `xml:"urn:vsan ConfigureDatastorePrincipal,omitempty"`
    Res *vsantypes.ConfigureDatastorePrincipalResponse `xml:"urn:vsan ConfigureDatastorePrincipalResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureDatastorePrincipalBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureDatastorePrincipal(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureDatastorePrincipal) (*vsantypes.ConfigureDatastorePrincipalResponse, error) {
  var reqBody, resBody ConfigureDatastorePrincipalBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureEvcMode_TaskBody struct{
    Req *vsantypes.ConfigureEvcMode_Task `xml:"urn:vsan ConfigureEvcMode_Task,omitempty"`
    Res *vsantypes.ConfigureEvcMode_TaskResponse `xml:"urn:vsan ConfigureEvcMode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureEvcMode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureEvcMode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureEvcMode_Task) (*vsantypes.ConfigureEvcMode_TaskResponse, error) {
  var reqBody, resBody ConfigureEvcMode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureHostCache_TaskBody struct{
    Req *vsantypes.ConfigureHostCache_Task `xml:"urn:vsan ConfigureHostCache_Task,omitempty"`
    Res *vsantypes.ConfigureHostCache_TaskResponse `xml:"urn:vsan ConfigureHostCache_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureHostCache_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureHostCache_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureHostCache_Task) (*vsantypes.ConfigureHostCache_TaskResponse, error) {
  var reqBody, resBody ConfigureHostCache_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureLicenseSourceBody struct{
    Req *vsantypes.ConfigureLicenseSource `xml:"urn:vsan ConfigureLicenseSource,omitempty"`
    Res *vsantypes.ConfigureLicenseSourceResponse `xml:"urn:vsan ConfigureLicenseSourceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureLicenseSourceBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureLicenseSource(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureLicenseSource) (*vsantypes.ConfigureLicenseSourceResponse, error) {
  var reqBody, resBody ConfigureLicenseSourceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigurePowerPolicyBody struct{
    Req *vsantypes.ConfigurePowerPolicy `xml:"urn:vsan ConfigurePowerPolicy,omitempty"`
    Res *vsantypes.ConfigurePowerPolicyResponse `xml:"urn:vsan ConfigurePowerPolicyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigurePowerPolicyBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigurePowerPolicy(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigurePowerPolicy) (*vsantypes.ConfigurePowerPolicyResponse, error) {
  var reqBody, resBody ConfigurePowerPolicyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureStorageDrsForPod_TaskBody struct{
    Req *vsantypes.ConfigureStorageDrsForPod_Task `xml:"urn:vsan ConfigureStorageDrsForPod_Task,omitempty"`
    Res *vsantypes.ConfigureStorageDrsForPod_TaskResponse `xml:"urn:vsan ConfigureStorageDrsForPod_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureStorageDrsForPod_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureStorageDrsForPod_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureStorageDrsForPod_Task) (*vsantypes.ConfigureStorageDrsForPod_TaskResponse, error) {
  var reqBody, resBody ConfigureStorageDrsForPod_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureVFlashResourceEx_TaskBody struct{
    Req *vsantypes.ConfigureVFlashResourceEx_Task `xml:"urn:vsan ConfigureVFlashResourceEx_Task,omitempty"`
    Res *vsantypes.ConfigureVFlashResourceEx_TaskResponse `xml:"urn:vsan ConfigureVFlashResourceEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureVFlashResourceEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureVFlashResourceEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureVFlashResourceEx_Task) (*vsantypes.ConfigureVFlashResourceEx_TaskResponse, error) {
  var reqBody, resBody ConfigureVFlashResourceEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConsolidateVMDisks_TaskBody struct{
    Req *vsantypes.ConsolidateVMDisks_Task `xml:"urn:vsan ConsolidateVMDisks_Task,omitempty"`
    Res *vsantypes.ConsolidateVMDisks_TaskResponse `xml:"urn:vsan ConsolidateVMDisks_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConsolidateVMDisks_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ConsolidateVMDisks_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConsolidateVMDisks_Task) (*vsantypes.ConsolidateVMDisks_TaskResponse, error) {
  var reqBody, resBody ConsolidateVMDisks_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ContinueRetrievePropertiesExBody struct{
    Req *vsantypes.ContinueRetrievePropertiesEx `xml:"urn:vsan ContinueRetrievePropertiesEx,omitempty"`
    Res *vsantypes.ContinueRetrievePropertiesExResponse `xml:"urn:vsan ContinueRetrievePropertiesExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ContinueRetrievePropertiesExBody) Fault() *soap.Fault { return b.Fault_ }

func ContinueRetrievePropertiesEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.ContinueRetrievePropertiesEx) (*vsantypes.ContinueRetrievePropertiesExResponse, error) {
  var reqBody, resBody ContinueRetrievePropertiesExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConvertNamespacePathToUuidPathBody struct{
    Req *vsantypes.ConvertNamespacePathToUuidPath `xml:"urn:vsan ConvertNamespacePathToUuidPath,omitempty"`
    Res *vsantypes.ConvertNamespacePathToUuidPathResponse `xml:"urn:vsan ConvertNamespacePathToUuidPathResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConvertNamespacePathToUuidPathBody) Fault() *soap.Fault { return b.Fault_ }

func ConvertNamespacePathToUuidPath(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConvertNamespacePathToUuidPath) (*vsantypes.ConvertNamespacePathToUuidPathResponse, error) {
  var reqBody, resBody ConvertNamespacePathToUuidPathBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CopyDatastoreFile_TaskBody struct{
    Req *vsantypes.CopyDatastoreFile_Task `xml:"urn:vsan CopyDatastoreFile_Task,omitempty"`
    Res *vsantypes.CopyDatastoreFile_TaskResponse `xml:"urn:vsan CopyDatastoreFile_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CopyDatastoreFile_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CopyDatastoreFile_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CopyDatastoreFile_Task) (*vsantypes.CopyDatastoreFile_TaskResponse, error) {
  var reqBody, resBody CopyDatastoreFile_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CopyVirtualDisk_TaskBody struct{
    Req *vsantypes.CopyVirtualDisk_Task `xml:"urn:vsan CopyVirtualDisk_Task,omitempty"`
    Res *vsantypes.CopyVirtualDisk_TaskResponse `xml:"urn:vsan CopyVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CopyVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CopyVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CopyVirtualDisk_Task) (*vsantypes.CopyVirtualDisk_TaskResponse, error) {
  var reqBody, resBody CopyVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateAlarmBody struct{
    Req *vsantypes.CreateAlarm `xml:"urn:vsan CreateAlarm,omitempty"`
    Res *vsantypes.CreateAlarmResponse `xml:"urn:vsan CreateAlarmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateAlarmBody) Fault() *soap.Fault { return b.Fault_ }

func CreateAlarm(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateAlarm) (*vsantypes.CreateAlarmResponse, error) {
  var reqBody, resBody CreateAlarmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateChildVM_TaskBody struct{
    Req *vsantypes.CreateChildVM_Task `xml:"urn:vsan CreateChildVM_Task,omitempty"`
    Res *vsantypes.CreateChildVM_TaskResponse `xml:"urn:vsan CreateChildVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateChildVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateChildVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateChildVM_Task) (*vsantypes.CreateChildVM_TaskResponse, error) {
  var reqBody, resBody CreateChildVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateClusterBody struct{
    Req *vsantypes.CreateCluster `xml:"urn:vsan CreateCluster,omitempty"`
    Res *vsantypes.CreateClusterResponse `xml:"urn:vsan CreateClusterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateClusterBody) Fault() *soap.Fault { return b.Fault_ }

func CreateCluster(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateCluster) (*vsantypes.CreateClusterResponse, error) {
  var reqBody, resBody CreateClusterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateClusterExBody struct{
    Req *vsantypes.CreateClusterEx `xml:"urn:vsan CreateClusterEx,omitempty"`
    Res *vsantypes.CreateClusterExResponse `xml:"urn:vsan CreateClusterExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateClusterExBody) Fault() *soap.Fault { return b.Fault_ }

func CreateClusterEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateClusterEx) (*vsantypes.CreateClusterExResponse, error) {
  var reqBody, resBody CreateClusterExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateCollectorForEventsBody struct{
    Req *vsantypes.CreateCollectorForEvents `xml:"urn:vsan CreateCollectorForEvents,omitempty"`
    Res *vsantypes.CreateCollectorForEventsResponse `xml:"urn:vsan CreateCollectorForEventsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateCollectorForEventsBody) Fault() *soap.Fault { return b.Fault_ }

func CreateCollectorForEvents(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateCollectorForEvents) (*vsantypes.CreateCollectorForEventsResponse, error) {
  var reqBody, resBody CreateCollectorForEventsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateCollectorForTasksBody struct{
    Req *vsantypes.CreateCollectorForTasks `xml:"urn:vsan CreateCollectorForTasks,omitempty"`
    Res *vsantypes.CreateCollectorForTasksResponse `xml:"urn:vsan CreateCollectorForTasksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateCollectorForTasksBody) Fault() *soap.Fault { return b.Fault_ }

func CreateCollectorForTasks(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateCollectorForTasks) (*vsantypes.CreateCollectorForTasksResponse, error) {
  var reqBody, resBody CreateCollectorForTasksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateContainerViewBody struct{
    Req *vsantypes.CreateContainerView `xml:"urn:vsan CreateContainerView,omitempty"`
    Res *vsantypes.CreateContainerViewResponse `xml:"urn:vsan CreateContainerViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateContainerViewBody) Fault() *soap.Fault { return b.Fault_ }

func CreateContainerView(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateContainerView) (*vsantypes.CreateContainerViewResponse, error) {
  var reqBody, resBody CreateContainerViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateCustomizationSpecBody struct{
    Req *vsantypes.CreateCustomizationSpec `xml:"urn:vsan CreateCustomizationSpec,omitempty"`
    Res *vsantypes.CreateCustomizationSpecResponse `xml:"urn:vsan CreateCustomizationSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateCustomizationSpecBody) Fault() *soap.Fault { return b.Fault_ }

func CreateCustomizationSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateCustomizationSpec) (*vsantypes.CreateCustomizationSpecResponse, error) {
  var reqBody, resBody CreateCustomizationSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDVPortgroup_TaskBody struct{
    Req *vsantypes.CreateDVPortgroup_Task `xml:"urn:vsan CreateDVPortgroup_Task,omitempty"`
    Res *vsantypes.CreateDVPortgroup_TaskResponse `xml:"urn:vsan CreateDVPortgroup_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDVPortgroup_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDVPortgroup_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDVPortgroup_Task) (*vsantypes.CreateDVPortgroup_TaskResponse, error) {
  var reqBody, resBody CreateDVPortgroup_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDVS_TaskBody struct{
    Req *vsantypes.CreateDVS_Task `xml:"urn:vsan CreateDVS_Task,omitempty"`
    Res *vsantypes.CreateDVS_TaskResponse `xml:"urn:vsan CreateDVS_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDVS_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDVS_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDVS_Task) (*vsantypes.CreateDVS_TaskResponse, error) {
  var reqBody, resBody CreateDVS_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDatacenterBody struct{
    Req *vsantypes.CreateDatacenter `xml:"urn:vsan CreateDatacenter,omitempty"`
    Res *vsantypes.CreateDatacenterResponse `xml:"urn:vsan CreateDatacenterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDatacenterBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDatacenter(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDatacenter) (*vsantypes.CreateDatacenterResponse, error) {
  var reqBody, resBody CreateDatacenterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDefaultProfileBody struct{
    Req *vsantypes.CreateDefaultProfile `xml:"urn:vsan CreateDefaultProfile,omitempty"`
    Res *vsantypes.CreateDefaultProfileResponse `xml:"urn:vsan CreateDefaultProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDefaultProfileBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDefaultProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDefaultProfile) (*vsantypes.CreateDefaultProfileResponse, error) {
  var reqBody, resBody CreateDefaultProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDescriptorBody struct{
    Req *vsantypes.CreateDescriptor `xml:"urn:vsan CreateDescriptor,omitempty"`
    Res *vsantypes.CreateDescriptorResponse `xml:"urn:vsan CreateDescriptorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDescriptorBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDescriptor(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDescriptor) (*vsantypes.CreateDescriptorResponse, error) {
  var reqBody, resBody CreateDescriptorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDiagnosticPartitionBody struct{
    Req *vsantypes.CreateDiagnosticPartition `xml:"urn:vsan CreateDiagnosticPartition,omitempty"`
    Res *vsantypes.CreateDiagnosticPartitionResponse `xml:"urn:vsan CreateDiagnosticPartitionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDiagnosticPartitionBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDiagnosticPartition(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDiagnosticPartition) (*vsantypes.CreateDiagnosticPartitionResponse, error) {
  var reqBody, resBody CreateDiagnosticPartitionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDirectoryBody struct{
    Req *vsantypes.CreateDirectory `xml:"urn:vsan CreateDirectory,omitempty"`
    Res *vsantypes.CreateDirectoryResponse `xml:"urn:vsan CreateDirectoryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDirectoryBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDirectory(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDirectory) (*vsantypes.CreateDirectoryResponse, error) {
  var reqBody, resBody CreateDirectoryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateDisk_TaskBody struct{
    Req *vsantypes.CreateDisk_Task `xml:"urn:vsan CreateDisk_Task,omitempty"`
    Res *vsantypes.CreateDisk_TaskResponse `xml:"urn:vsan CreateDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateDisk_Task) (*vsantypes.CreateDisk_TaskResponse, error) {
  var reqBody, resBody CreateDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateFilterBody struct{
    Req *vsantypes.CreateFilter `xml:"urn:vsan CreateFilter,omitempty"`
    Res *vsantypes.CreateFilterResponse `xml:"urn:vsan CreateFilterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateFilterBody) Fault() *soap.Fault { return b.Fault_ }

func CreateFilter(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateFilter) (*vsantypes.CreateFilterResponse, error) {
  var reqBody, resBody CreateFilterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateFolderBody struct{
    Req *vsantypes.CreateFolder `xml:"urn:vsan CreateFolder,omitempty"`
    Res *vsantypes.CreateFolderResponse `xml:"urn:vsan CreateFolderResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateFolderBody) Fault() *soap.Fault { return b.Fault_ }

func CreateFolder(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateFolder) (*vsantypes.CreateFolderResponse, error) {
  var reqBody, resBody CreateFolderBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateGroupBody struct{
    Req *vsantypes.CreateGroup `xml:"urn:vsan CreateGroup,omitempty"`
    Res *vsantypes.CreateGroupResponse `xml:"urn:vsan CreateGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateGroupBody) Fault() *soap.Fault { return b.Fault_ }

func CreateGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateGroup) (*vsantypes.CreateGroupResponse, error) {
  var reqBody, resBody CreateGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateImportSpecBody struct{
    Req *vsantypes.CreateImportSpec `xml:"urn:vsan CreateImportSpec,omitempty"`
    Res *vsantypes.CreateImportSpecResponse `xml:"urn:vsan CreateImportSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateImportSpecBody) Fault() *soap.Fault { return b.Fault_ }

func CreateImportSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateImportSpec) (*vsantypes.CreateImportSpecResponse, error) {
  var reqBody, resBody CreateImportSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateInventoryViewBody struct{
    Req *vsantypes.CreateInventoryView `xml:"urn:vsan CreateInventoryView,omitempty"`
    Res *vsantypes.CreateInventoryViewResponse `xml:"urn:vsan CreateInventoryViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateInventoryViewBody) Fault() *soap.Fault { return b.Fault_ }

func CreateInventoryView(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateInventoryView) (*vsantypes.CreateInventoryViewResponse, error) {
  var reqBody, resBody CreateInventoryViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateIpPoolBody struct{
    Req *vsantypes.CreateIpPool `xml:"urn:vsan CreateIpPool,omitempty"`
    Res *vsantypes.CreateIpPoolResponse `xml:"urn:vsan CreateIpPoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateIpPoolBody) Fault() *soap.Fault { return b.Fault_ }

func CreateIpPool(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateIpPool) (*vsantypes.CreateIpPoolResponse, error) {
  var reqBody, resBody CreateIpPoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateListViewBody struct{
    Req *vsantypes.CreateListView `xml:"urn:vsan CreateListView,omitempty"`
    Res *vsantypes.CreateListViewResponse `xml:"urn:vsan CreateListViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateListViewBody) Fault() *soap.Fault { return b.Fault_ }

func CreateListView(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateListView) (*vsantypes.CreateListViewResponse, error) {
  var reqBody, resBody CreateListViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateListViewFromViewBody struct{
    Req *vsantypes.CreateListViewFromView `xml:"urn:vsan CreateListViewFromView,omitempty"`
    Res *vsantypes.CreateListViewFromViewResponse `xml:"urn:vsan CreateListViewFromViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateListViewFromViewBody) Fault() *soap.Fault { return b.Fault_ }

func CreateListViewFromView(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateListViewFromView) (*vsantypes.CreateListViewFromViewResponse, error) {
  var reqBody, resBody CreateListViewFromViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateLocalDatastoreBody struct{
    Req *vsantypes.CreateLocalDatastore `xml:"urn:vsan CreateLocalDatastore,omitempty"`
    Res *vsantypes.CreateLocalDatastoreResponse `xml:"urn:vsan CreateLocalDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateLocalDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func CreateLocalDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateLocalDatastore) (*vsantypes.CreateLocalDatastoreResponse, error) {
  var reqBody, resBody CreateLocalDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateNasDatastoreBody struct{
    Req *vsantypes.CreateNasDatastore `xml:"urn:vsan CreateNasDatastore,omitempty"`
    Res *vsantypes.CreateNasDatastoreResponse `xml:"urn:vsan CreateNasDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateNasDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func CreateNasDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateNasDatastore) (*vsantypes.CreateNasDatastoreResponse, error) {
  var reqBody, resBody CreateNasDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateObjectScheduledTaskBody struct{
    Req *vsantypes.CreateObjectScheduledTask `xml:"urn:vsan CreateObjectScheduledTask,omitempty"`
    Res *vsantypes.CreateObjectScheduledTaskResponse `xml:"urn:vsan CreateObjectScheduledTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateObjectScheduledTaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateObjectScheduledTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateObjectScheduledTask) (*vsantypes.CreateObjectScheduledTaskResponse, error) {
  var reqBody, resBody CreateObjectScheduledTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreatePerfIntervalBody struct{
    Req *vsantypes.CreatePerfInterval `xml:"urn:vsan CreatePerfInterval,omitempty"`
    Res *vsantypes.CreatePerfIntervalResponse `xml:"urn:vsan CreatePerfIntervalResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreatePerfIntervalBody) Fault() *soap.Fault { return b.Fault_ }

func CreatePerfInterval(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreatePerfInterval) (*vsantypes.CreatePerfIntervalResponse, error) {
  var reqBody, resBody CreatePerfIntervalBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateProfileBody struct{
    Req *vsantypes.CreateProfile `xml:"urn:vsan CreateProfile,omitempty"`
    Res *vsantypes.CreateProfileResponse `xml:"urn:vsan CreateProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateProfileBody) Fault() *soap.Fault { return b.Fault_ }

func CreateProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateProfile) (*vsantypes.CreateProfileResponse, error) {
  var reqBody, resBody CreateProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreatePropertyCollectorBody struct{
    Req *vsantypes.CreatePropertyCollector `xml:"urn:vsan CreatePropertyCollector,omitempty"`
    Res *vsantypes.CreatePropertyCollectorResponse `xml:"urn:vsan CreatePropertyCollectorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreatePropertyCollectorBody) Fault() *soap.Fault { return b.Fault_ }

func CreatePropertyCollector(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreatePropertyCollector) (*vsantypes.CreatePropertyCollectorResponse, error) {
  var reqBody, resBody CreatePropertyCollectorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateRegistryKeyInGuestBody struct{
    Req *vsantypes.CreateRegistryKeyInGuest `xml:"urn:vsan CreateRegistryKeyInGuest,omitempty"`
    Res *vsantypes.CreateRegistryKeyInGuestResponse `xml:"urn:vsan CreateRegistryKeyInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateRegistryKeyInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func CreateRegistryKeyInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateRegistryKeyInGuest) (*vsantypes.CreateRegistryKeyInGuestResponse, error) {
  var reqBody, resBody CreateRegistryKeyInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateResourcePoolBody struct{
    Req *vsantypes.CreateResourcePool `xml:"urn:vsan CreateResourcePool,omitempty"`
    Res *vsantypes.CreateResourcePoolResponse `xml:"urn:vsan CreateResourcePoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateResourcePoolBody) Fault() *soap.Fault { return b.Fault_ }

func CreateResourcePool(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateResourcePool) (*vsantypes.CreateResourcePoolResponse, error) {
  var reqBody, resBody CreateResourcePoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateScheduledTaskBody struct{
    Req *vsantypes.CreateScheduledTask `xml:"urn:vsan CreateScheduledTask,omitempty"`
    Res *vsantypes.CreateScheduledTaskResponse `xml:"urn:vsan CreateScheduledTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateScheduledTaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateScheduledTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateScheduledTask) (*vsantypes.CreateScheduledTaskResponse, error) {
  var reqBody, resBody CreateScheduledTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateScreenshot_TaskBody struct{
    Req *vsantypes.CreateScreenshot_Task `xml:"urn:vsan CreateScreenshot_Task,omitempty"`
    Res *vsantypes.CreateScreenshot_TaskResponse `xml:"urn:vsan CreateScreenshot_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateScreenshot_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateScreenshot_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateScreenshot_Task) (*vsantypes.CreateScreenshot_TaskResponse, error) {
  var reqBody, resBody CreateScreenshot_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateSecondaryVMEx_TaskBody struct{
    Req *vsantypes.CreateSecondaryVMEx_Task `xml:"urn:vsan CreateSecondaryVMEx_Task,omitempty"`
    Res *vsantypes.CreateSecondaryVMEx_TaskResponse `xml:"urn:vsan CreateSecondaryVMEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateSecondaryVMEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateSecondaryVMEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateSecondaryVMEx_Task) (*vsantypes.CreateSecondaryVMEx_TaskResponse, error) {
  var reqBody, resBody CreateSecondaryVMEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateSecondaryVM_TaskBody struct{
    Req *vsantypes.CreateSecondaryVM_Task `xml:"urn:vsan CreateSecondaryVM_Task,omitempty"`
    Res *vsantypes.CreateSecondaryVM_TaskResponse `xml:"urn:vsan CreateSecondaryVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateSecondaryVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateSecondaryVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateSecondaryVM_Task) (*vsantypes.CreateSecondaryVM_TaskResponse, error) {
  var reqBody, resBody CreateSecondaryVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateSnapshotEx_TaskBody struct{
    Req *vsantypes.CreateSnapshotEx_Task `xml:"urn:vsan CreateSnapshotEx_Task,omitempty"`
    Res *vsantypes.CreateSnapshotEx_TaskResponse `xml:"urn:vsan CreateSnapshotEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateSnapshotEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateSnapshotEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateSnapshotEx_Task) (*vsantypes.CreateSnapshotEx_TaskResponse, error) {
  var reqBody, resBody CreateSnapshotEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateSnapshot_TaskBody struct{
    Req *vsantypes.CreateSnapshot_Task `xml:"urn:vsan CreateSnapshot_Task,omitempty"`
    Res *vsantypes.CreateSnapshot_TaskResponse `xml:"urn:vsan CreateSnapshot_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateSnapshot_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateSnapshot_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateSnapshot_Task) (*vsantypes.CreateSnapshot_TaskResponse, error) {
  var reqBody, resBody CreateSnapshot_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateStoragePodBody struct{
    Req *vsantypes.CreateStoragePod `xml:"urn:vsan CreateStoragePod,omitempty"`
    Res *vsantypes.CreateStoragePodResponse `xml:"urn:vsan CreateStoragePodResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateStoragePodBody) Fault() *soap.Fault { return b.Fault_ }

func CreateStoragePod(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateStoragePod) (*vsantypes.CreateStoragePodResponse, error) {
  var reqBody, resBody CreateStoragePodBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateTaskBody struct{
    Req *vsantypes.CreateTask `xml:"urn:vsan CreateTask,omitempty"`
    Res *vsantypes.CreateTaskResponse `xml:"urn:vsan CreateTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateTaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateTask) (*vsantypes.CreateTaskResponse, error) {
  var reqBody, resBody CreateTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateTemporaryDirectoryInGuestBody struct{
    Req *vsantypes.CreateTemporaryDirectoryInGuest `xml:"urn:vsan CreateTemporaryDirectoryInGuest,omitempty"`
    Res *vsantypes.CreateTemporaryDirectoryInGuestResponse `xml:"urn:vsan CreateTemporaryDirectoryInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateTemporaryDirectoryInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func CreateTemporaryDirectoryInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateTemporaryDirectoryInGuest) (*vsantypes.CreateTemporaryDirectoryInGuestResponse, error) {
  var reqBody, resBody CreateTemporaryDirectoryInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateTemporaryFileInGuestBody struct{
    Req *vsantypes.CreateTemporaryFileInGuest `xml:"urn:vsan CreateTemporaryFileInGuest,omitempty"`
    Res *vsantypes.CreateTemporaryFileInGuestResponse `xml:"urn:vsan CreateTemporaryFileInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateTemporaryFileInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func CreateTemporaryFileInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateTemporaryFileInGuest) (*vsantypes.CreateTemporaryFileInGuestResponse, error) {
  var reqBody, resBody CreateTemporaryFileInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateUserBody struct{
    Req *vsantypes.CreateUser `xml:"urn:vsan CreateUser,omitempty"`
    Res *vsantypes.CreateUserResponse `xml:"urn:vsan CreateUserResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateUserBody) Fault() *soap.Fault { return b.Fault_ }

func CreateUser(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateUser) (*vsantypes.CreateUserResponse, error) {
  var reqBody, resBody CreateUserBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateVAppBody struct{
    Req *vsantypes.CreateVApp `xml:"urn:vsan CreateVApp,omitempty"`
    Res *vsantypes.CreateVAppResponse `xml:"urn:vsan CreateVAppResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateVAppBody) Fault() *soap.Fault { return b.Fault_ }

func CreateVApp(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateVApp) (*vsantypes.CreateVAppResponse, error) {
  var reqBody, resBody CreateVAppBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateVM_TaskBody struct{
    Req *vsantypes.CreateVM_Task `xml:"urn:vsan CreateVM_Task,omitempty"`
    Res *vsantypes.CreateVM_TaskResponse `xml:"urn:vsan CreateVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateVM_Task) (*vsantypes.CreateVM_TaskResponse, error) {
  var reqBody, resBody CreateVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateVirtualDisk_TaskBody struct{
    Req *vsantypes.CreateVirtualDisk_Task `xml:"urn:vsan CreateVirtualDisk_Task,omitempty"`
    Res *vsantypes.CreateVirtualDisk_TaskResponse `xml:"urn:vsan CreateVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateVirtualDisk_Task) (*vsantypes.CreateVirtualDisk_TaskResponse, error) {
  var reqBody, resBody CreateVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateVmfsDatastoreBody struct{
    Req *vsantypes.CreateVmfsDatastore `xml:"urn:vsan CreateVmfsDatastore,omitempty"`
    Res *vsantypes.CreateVmfsDatastoreResponse `xml:"urn:vsan CreateVmfsDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateVmfsDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func CreateVmfsDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateVmfsDatastore) (*vsantypes.CreateVmfsDatastoreResponse, error) {
  var reqBody, resBody CreateVmfsDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateVvolDatastoreBody struct{
    Req *vsantypes.CreateVvolDatastore `xml:"urn:vsan CreateVvolDatastore,omitempty"`
    Res *vsantypes.CreateVvolDatastoreResponse `xml:"urn:vsan CreateVvolDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateVvolDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func CreateVvolDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateVvolDatastore) (*vsantypes.CreateVvolDatastoreResponse, error) {
  var reqBody, resBody CreateVvolDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CurrentTimeBody struct{
    Req *vsantypes.CurrentTime `xml:"urn:vsan CurrentTime,omitempty"`
    Res *vsantypes.CurrentTimeResponse `xml:"urn:vsan CurrentTimeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CurrentTimeBody) Fault() *soap.Fault { return b.Fault_ }

func CurrentTime(ctx context.Context, r soap.RoundTripper, req *vsantypes.CurrentTime) (*vsantypes.CurrentTimeResponse, error) {
  var reqBody, resBody CurrentTimeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CustomizationSpecItemToXmlBody struct{
    Req *vsantypes.CustomizationSpecItemToXml `xml:"urn:vsan CustomizationSpecItemToXml,omitempty"`
    Res *vsantypes.CustomizationSpecItemToXmlResponse `xml:"urn:vsan CustomizationSpecItemToXmlResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CustomizationSpecItemToXmlBody) Fault() *soap.Fault { return b.Fault_ }

func CustomizationSpecItemToXml(ctx context.Context, r soap.RoundTripper, req *vsantypes.CustomizationSpecItemToXml) (*vsantypes.CustomizationSpecItemToXmlResponse, error) {
  var reqBody, resBody CustomizationSpecItemToXmlBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CustomizeVM_TaskBody struct{
    Req *vsantypes.CustomizeVM_Task `xml:"urn:vsan CustomizeVM_Task,omitempty"`
    Res *vsantypes.CustomizeVM_TaskResponse `xml:"urn:vsan CustomizeVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CustomizeVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CustomizeVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CustomizeVM_Task) (*vsantypes.CustomizeVM_TaskResponse, error) {
  var reqBody, resBody CustomizeVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DVPortgroupRollback_TaskBody struct{
    Req *vsantypes.DVPortgroupRollback_Task `xml:"urn:vsan DVPortgroupRollback_Task,omitempty"`
    Res *vsantypes.DVPortgroupRollback_TaskResponse `xml:"urn:vsan DVPortgroupRollback_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DVPortgroupRollback_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DVPortgroupRollback_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DVPortgroupRollback_Task) (*vsantypes.DVPortgroupRollback_TaskResponse, error) {
  var reqBody, resBody DVPortgroupRollback_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DVSManagerExportEntity_TaskBody struct{
    Req *vsantypes.DVSManagerExportEntity_Task `xml:"urn:vsan DVSManagerExportEntity_Task,omitempty"`
    Res *vsantypes.DVSManagerExportEntity_TaskResponse `xml:"urn:vsan DVSManagerExportEntity_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DVSManagerExportEntity_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DVSManagerExportEntity_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DVSManagerExportEntity_Task) (*vsantypes.DVSManagerExportEntity_TaskResponse, error) {
  var reqBody, resBody DVSManagerExportEntity_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DVSManagerImportEntity_TaskBody struct{
    Req *vsantypes.DVSManagerImportEntity_Task `xml:"urn:vsan DVSManagerImportEntity_Task,omitempty"`
    Res *vsantypes.DVSManagerImportEntity_TaskResponse `xml:"urn:vsan DVSManagerImportEntity_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DVSManagerImportEntity_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DVSManagerImportEntity_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DVSManagerImportEntity_Task) (*vsantypes.DVSManagerImportEntity_TaskResponse, error) {
  var reqBody, resBody DVSManagerImportEntity_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DVSManagerLookupDvPortGroupBody struct{
    Req *vsantypes.DVSManagerLookupDvPortGroup `xml:"urn:vsan DVSManagerLookupDvPortGroup,omitempty"`
    Res *vsantypes.DVSManagerLookupDvPortGroupResponse `xml:"urn:vsan DVSManagerLookupDvPortGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DVSManagerLookupDvPortGroupBody) Fault() *soap.Fault { return b.Fault_ }

func DVSManagerLookupDvPortGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.DVSManagerLookupDvPortGroup) (*vsantypes.DVSManagerLookupDvPortGroupResponse, error) {
  var reqBody, resBody DVSManagerLookupDvPortGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DVSRollback_TaskBody struct{
    Req *vsantypes.DVSRollback_Task `xml:"urn:vsan DVSRollback_Task,omitempty"`
    Res *vsantypes.DVSRollback_TaskResponse `xml:"urn:vsan DVSRollback_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DVSRollback_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DVSRollback_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DVSRollback_Task) (*vsantypes.DVSRollback_TaskResponse, error) {
  var reqBody, resBody DVSRollback_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DatastoreEnterMaintenanceModeBody struct{
    Req *vsantypes.DatastoreEnterMaintenanceMode `xml:"urn:vsan DatastoreEnterMaintenanceMode,omitempty"`
    Res *vsantypes.DatastoreEnterMaintenanceModeResponse `xml:"urn:vsan DatastoreEnterMaintenanceModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DatastoreEnterMaintenanceModeBody) Fault() *soap.Fault { return b.Fault_ }

func DatastoreEnterMaintenanceMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.DatastoreEnterMaintenanceMode) (*vsantypes.DatastoreEnterMaintenanceModeResponse, error) {
  var reqBody, resBody DatastoreEnterMaintenanceModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DatastoreExitMaintenanceMode_TaskBody struct{
    Req *vsantypes.DatastoreExitMaintenanceMode_Task `xml:"urn:vsan DatastoreExitMaintenanceMode_Task,omitempty"`
    Res *vsantypes.DatastoreExitMaintenanceMode_TaskResponse `xml:"urn:vsan DatastoreExitMaintenanceMode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DatastoreExitMaintenanceMode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DatastoreExitMaintenanceMode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DatastoreExitMaintenanceMode_Task) (*vsantypes.DatastoreExitMaintenanceMode_TaskResponse, error) {
  var reqBody, resBody DatastoreExitMaintenanceMode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DecodeLicenseBody struct{
    Req *vsantypes.DecodeLicense `xml:"urn:vsan DecodeLicense,omitempty"`
    Res *vsantypes.DecodeLicenseResponse `xml:"urn:vsan DecodeLicenseResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DecodeLicenseBody) Fault() *soap.Fault { return b.Fault_ }

func DecodeLicense(ctx context.Context, r soap.RoundTripper, req *vsantypes.DecodeLicense) (*vsantypes.DecodeLicenseResponse, error) {
  var reqBody, resBody DecodeLicenseBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DefragmentAllDisksBody struct{
    Req *vsantypes.DefragmentAllDisks `xml:"urn:vsan DefragmentAllDisks,omitempty"`
    Res *vsantypes.DefragmentAllDisksResponse `xml:"urn:vsan DefragmentAllDisksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DefragmentAllDisksBody) Fault() *soap.Fault { return b.Fault_ }

func DefragmentAllDisks(ctx context.Context, r soap.RoundTripper, req *vsantypes.DefragmentAllDisks) (*vsantypes.DefragmentAllDisksResponse, error) {
  var reqBody, resBody DefragmentAllDisksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DefragmentVirtualDisk_TaskBody struct{
    Req *vsantypes.DefragmentVirtualDisk_Task `xml:"urn:vsan DefragmentVirtualDisk_Task,omitempty"`
    Res *vsantypes.DefragmentVirtualDisk_TaskResponse `xml:"urn:vsan DefragmentVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DefragmentVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DefragmentVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DefragmentVirtualDisk_Task) (*vsantypes.DefragmentVirtualDisk_TaskResponse, error) {
  var reqBody, resBody DefragmentVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteCustomizationSpecBody struct{
    Req *vsantypes.DeleteCustomizationSpec `xml:"urn:vsan DeleteCustomizationSpec,omitempty"`
    Res *vsantypes.DeleteCustomizationSpecResponse `xml:"urn:vsan DeleteCustomizationSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteCustomizationSpecBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteCustomizationSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteCustomizationSpec) (*vsantypes.DeleteCustomizationSpecResponse, error) {
  var reqBody, resBody DeleteCustomizationSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteDatastoreFile_TaskBody struct{
    Req *vsantypes.DeleteDatastoreFile_Task `xml:"urn:vsan DeleteDatastoreFile_Task,omitempty"`
    Res *vsantypes.DeleteDatastoreFile_TaskResponse `xml:"urn:vsan DeleteDatastoreFile_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteDatastoreFile_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteDatastoreFile_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteDatastoreFile_Task) (*vsantypes.DeleteDatastoreFile_TaskResponse, error) {
  var reqBody, resBody DeleteDatastoreFile_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteDirectoryBody struct{
    Req *vsantypes.DeleteDirectory `xml:"urn:vsan DeleteDirectory,omitempty"`
    Res *vsantypes.DeleteDirectoryResponse `xml:"urn:vsan DeleteDirectoryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteDirectoryBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteDirectory(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteDirectory) (*vsantypes.DeleteDirectoryResponse, error) {
  var reqBody, resBody DeleteDirectoryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteDirectoryInGuestBody struct{
    Req *vsantypes.DeleteDirectoryInGuest `xml:"urn:vsan DeleteDirectoryInGuest,omitempty"`
    Res *vsantypes.DeleteDirectoryInGuestResponse `xml:"urn:vsan DeleteDirectoryInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteDirectoryInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteDirectoryInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteDirectoryInGuest) (*vsantypes.DeleteDirectoryInGuestResponse, error) {
  var reqBody, resBody DeleteDirectoryInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteFileBody struct{
    Req *vsantypes.DeleteFile `xml:"urn:vsan DeleteFile,omitempty"`
    Res *vsantypes.DeleteFileResponse `xml:"urn:vsan DeleteFileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteFileBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteFile(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteFile) (*vsantypes.DeleteFileResponse, error) {
  var reqBody, resBody DeleteFileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteFileInGuestBody struct{
    Req *vsantypes.DeleteFileInGuest `xml:"urn:vsan DeleteFileInGuest,omitempty"`
    Res *vsantypes.DeleteFileInGuestResponse `xml:"urn:vsan DeleteFileInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteFileInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteFileInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteFileInGuest) (*vsantypes.DeleteFileInGuestResponse, error) {
  var reqBody, resBody DeleteFileInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteHostSpecificationBody struct{
    Req *vsantypes.DeleteHostSpecification `xml:"urn:vsan DeleteHostSpecification,omitempty"`
    Res *vsantypes.DeleteHostSpecificationResponse `xml:"urn:vsan DeleteHostSpecificationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteHostSpecificationBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteHostSpecification(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteHostSpecification) (*vsantypes.DeleteHostSpecificationResponse, error) {
  var reqBody, resBody DeleteHostSpecificationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteHostSubSpecificationBody struct{
    Req *vsantypes.DeleteHostSubSpecification `xml:"urn:vsan DeleteHostSubSpecification,omitempty"`
    Res *vsantypes.DeleteHostSubSpecificationResponse `xml:"urn:vsan DeleteHostSubSpecificationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteHostSubSpecificationBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteHostSubSpecification(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteHostSubSpecification) (*vsantypes.DeleteHostSubSpecificationResponse, error) {
  var reqBody, resBody DeleteHostSubSpecificationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteRegistryKeyInGuestBody struct{
    Req *vsantypes.DeleteRegistryKeyInGuest `xml:"urn:vsan DeleteRegistryKeyInGuest,omitempty"`
    Res *vsantypes.DeleteRegistryKeyInGuestResponse `xml:"urn:vsan DeleteRegistryKeyInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteRegistryKeyInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteRegistryKeyInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteRegistryKeyInGuest) (*vsantypes.DeleteRegistryKeyInGuestResponse, error) {
  var reqBody, resBody DeleteRegistryKeyInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteRegistryValueInGuestBody struct{
    Req *vsantypes.DeleteRegistryValueInGuest `xml:"urn:vsan DeleteRegistryValueInGuest,omitempty"`
    Res *vsantypes.DeleteRegistryValueInGuestResponse `xml:"urn:vsan DeleteRegistryValueInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteRegistryValueInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteRegistryValueInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteRegistryValueInGuest) (*vsantypes.DeleteRegistryValueInGuestResponse, error) {
  var reqBody, resBody DeleteRegistryValueInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteScsiLunStateBody struct{
    Req *vsantypes.DeleteScsiLunState `xml:"urn:vsan DeleteScsiLunState,omitempty"`
    Res *vsantypes.DeleteScsiLunStateResponse `xml:"urn:vsan DeleteScsiLunStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteScsiLunStateBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteScsiLunState(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteScsiLunState) (*vsantypes.DeleteScsiLunStateResponse, error) {
  var reqBody, resBody DeleteScsiLunStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteVStorageObject_TaskBody struct{
    Req *vsantypes.DeleteVStorageObject_Task `xml:"urn:vsan DeleteVStorageObject_Task,omitempty"`
    Res *vsantypes.DeleteVStorageObject_TaskResponse `xml:"urn:vsan DeleteVStorageObject_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteVStorageObject_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteVStorageObject_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteVStorageObject_Task) (*vsantypes.DeleteVStorageObject_TaskResponse, error) {
  var reqBody, resBody DeleteVStorageObject_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteVffsVolumeStateBody struct{
    Req *vsantypes.DeleteVffsVolumeState `xml:"urn:vsan DeleteVffsVolumeState,omitempty"`
    Res *vsantypes.DeleteVffsVolumeStateResponse `xml:"urn:vsan DeleteVffsVolumeStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteVffsVolumeStateBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteVffsVolumeState(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteVffsVolumeState) (*vsantypes.DeleteVffsVolumeStateResponse, error) {
  var reqBody, resBody DeleteVffsVolumeStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteVirtualDisk_TaskBody struct{
    Req *vsantypes.DeleteVirtualDisk_Task `xml:"urn:vsan DeleteVirtualDisk_Task,omitempty"`
    Res *vsantypes.DeleteVirtualDisk_TaskResponse `xml:"urn:vsan DeleteVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteVirtualDisk_Task) (*vsantypes.DeleteVirtualDisk_TaskResponse, error) {
  var reqBody, resBody DeleteVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteVmfsVolumeStateBody struct{
    Req *vsantypes.DeleteVmfsVolumeState `xml:"urn:vsan DeleteVmfsVolumeState,omitempty"`
    Res *vsantypes.DeleteVmfsVolumeStateResponse `xml:"urn:vsan DeleteVmfsVolumeStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteVmfsVolumeStateBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteVmfsVolumeState(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteVmfsVolumeState) (*vsantypes.DeleteVmfsVolumeStateResponse, error) {
  var reqBody, resBody DeleteVmfsVolumeStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeleteVsanObjectsBody struct{
    Req *vsantypes.DeleteVsanObjects `xml:"urn:vsan DeleteVsanObjects,omitempty"`
    Res *vsantypes.DeleteVsanObjectsResponse `xml:"urn:vsan DeleteVsanObjectsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeleteVsanObjectsBody) Fault() *soap.Fault { return b.Fault_ }

func DeleteVsanObjects(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeleteVsanObjects) (*vsantypes.DeleteVsanObjectsResponse, error) {
  var reqBody, resBody DeleteVsanObjectsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeselectVnicBody struct{
    Req *vsantypes.DeselectVnic `xml:"urn:vsan DeselectVnic,omitempty"`
    Res *vsantypes.DeselectVnicResponse `xml:"urn:vsan DeselectVnicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeselectVnicBody) Fault() *soap.Fault { return b.Fault_ }

func DeselectVnic(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeselectVnic) (*vsantypes.DeselectVnicResponse, error) {
  var reqBody, resBody DeselectVnicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeselectVnicForNicTypeBody struct{
    Req *vsantypes.DeselectVnicForNicType `xml:"urn:vsan DeselectVnicForNicType,omitempty"`
    Res *vsantypes.DeselectVnicForNicTypeResponse `xml:"urn:vsan DeselectVnicForNicTypeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeselectVnicForNicTypeBody) Fault() *soap.Fault { return b.Fault_ }

func DeselectVnicForNicType(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeselectVnicForNicType) (*vsantypes.DeselectVnicForNicTypeResponse, error) {
  var reqBody, resBody DeselectVnicForNicTypeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyChildrenBody struct{
    Req *vsantypes.DestroyChildren `xml:"urn:vsan DestroyChildren,omitempty"`
    Res *vsantypes.DestroyChildrenResponse `xml:"urn:vsan DestroyChildrenResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyChildrenBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyChildren(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyChildren) (*vsantypes.DestroyChildrenResponse, error) {
  var reqBody, resBody DestroyChildrenBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyCollectorBody struct{
    Req *vsantypes.DestroyCollector `xml:"urn:vsan DestroyCollector,omitempty"`
    Res *vsantypes.DestroyCollectorResponse `xml:"urn:vsan DestroyCollectorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyCollectorBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyCollector(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyCollector) (*vsantypes.DestroyCollectorResponse, error) {
  var reqBody, resBody DestroyCollectorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyDatastoreBody struct{
    Req *vsantypes.DestroyDatastore `xml:"urn:vsan DestroyDatastore,omitempty"`
    Res *vsantypes.DestroyDatastoreResponse `xml:"urn:vsan DestroyDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyDatastore) (*vsantypes.DestroyDatastoreResponse, error) {
  var reqBody, resBody DestroyDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyIpPoolBody struct{
    Req *vsantypes.DestroyIpPool `xml:"urn:vsan DestroyIpPool,omitempty"`
    Res *vsantypes.DestroyIpPoolResponse `xml:"urn:vsan DestroyIpPoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyIpPoolBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyIpPool(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyIpPool) (*vsantypes.DestroyIpPoolResponse, error) {
  var reqBody, resBody DestroyIpPoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyNetworkBody struct{
    Req *vsantypes.DestroyNetwork `xml:"urn:vsan DestroyNetwork,omitempty"`
    Res *vsantypes.DestroyNetworkResponse `xml:"urn:vsan DestroyNetworkResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyNetworkBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyNetwork(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyNetwork) (*vsantypes.DestroyNetworkResponse, error) {
  var reqBody, resBody DestroyNetworkBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyProfileBody struct{
    Req *vsantypes.DestroyProfile `xml:"urn:vsan DestroyProfile,omitempty"`
    Res *vsantypes.DestroyProfileResponse `xml:"urn:vsan DestroyProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyProfileBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyProfile) (*vsantypes.DestroyProfileResponse, error) {
  var reqBody, resBody DestroyProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyPropertyCollectorBody struct{
    Req *vsantypes.DestroyPropertyCollector `xml:"urn:vsan DestroyPropertyCollector,omitempty"`
    Res *vsantypes.DestroyPropertyCollectorResponse `xml:"urn:vsan DestroyPropertyCollectorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyPropertyCollectorBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyPropertyCollector(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyPropertyCollector) (*vsantypes.DestroyPropertyCollectorResponse, error) {
  var reqBody, resBody DestroyPropertyCollectorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyPropertyFilterBody struct{
    Req *vsantypes.DestroyPropertyFilter `xml:"urn:vsan DestroyPropertyFilter,omitempty"`
    Res *vsantypes.DestroyPropertyFilterResponse `xml:"urn:vsan DestroyPropertyFilterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyPropertyFilterBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyPropertyFilter(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyPropertyFilter) (*vsantypes.DestroyPropertyFilterResponse, error) {
  var reqBody, resBody DestroyPropertyFilterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyVffsBody struct{
    Req *vsantypes.DestroyVffs `xml:"urn:vsan DestroyVffs,omitempty"`
    Res *vsantypes.DestroyVffsResponse `xml:"urn:vsan DestroyVffsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyVffsBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyVffs(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyVffs) (*vsantypes.DestroyVffsResponse, error) {
  var reqBody, resBody DestroyVffsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyViewBody struct{
    Req *vsantypes.DestroyView `xml:"urn:vsan DestroyView,omitempty"`
    Res *vsantypes.DestroyViewResponse `xml:"urn:vsan DestroyViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyViewBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyView(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyView) (*vsantypes.DestroyViewResponse, error) {
  var reqBody, resBody DestroyViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type Destroy_TaskBody struct{
    Req *vsantypes.Destroy_Task `xml:"urn:vsan Destroy_Task,omitempty"`
    Res *vsantypes.Destroy_TaskResponse `xml:"urn:vsan Destroy_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *Destroy_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func Destroy_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.Destroy_Task) (*vsantypes.Destroy_TaskResponse, error) {
  var reqBody, resBody Destroy_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DetachDisk_TaskBody struct{
    Req *vsantypes.DetachDisk_Task `xml:"urn:vsan DetachDisk_Task,omitempty"`
    Res *vsantypes.DetachDisk_TaskResponse `xml:"urn:vsan DetachDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DetachDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DetachDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DetachDisk_Task) (*vsantypes.DetachDisk_TaskResponse, error) {
  var reqBody, resBody DetachDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DetachScsiLunBody struct{
    Req *vsantypes.DetachScsiLun `xml:"urn:vsan DetachScsiLun,omitempty"`
    Res *vsantypes.DetachScsiLunResponse `xml:"urn:vsan DetachScsiLunResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DetachScsiLunBody) Fault() *soap.Fault { return b.Fault_ }

func DetachScsiLun(ctx context.Context, r soap.RoundTripper, req *vsantypes.DetachScsiLun) (*vsantypes.DetachScsiLunResponse, error) {
  var reqBody, resBody DetachScsiLunBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DetachScsiLunEx_TaskBody struct{
    Req *vsantypes.DetachScsiLunEx_Task `xml:"urn:vsan DetachScsiLunEx_Task,omitempty"`
    Res *vsantypes.DetachScsiLunEx_TaskResponse `xml:"urn:vsan DetachScsiLunEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DetachScsiLunEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DetachScsiLunEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DetachScsiLunEx_Task) (*vsantypes.DetachScsiLunEx_TaskResponse, error) {
  var reqBody, resBody DetachScsiLunEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DetachTagFromVStorageObjectBody struct{
    Req *vsantypes.DetachTagFromVStorageObject `xml:"urn:vsan DetachTagFromVStorageObject,omitempty"`
    Res *vsantypes.DetachTagFromVStorageObjectResponse `xml:"urn:vsan DetachTagFromVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DetachTagFromVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func DetachTagFromVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.DetachTagFromVStorageObject) (*vsantypes.DetachTagFromVStorageObjectResponse, error) {
  var reqBody, resBody DetachTagFromVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisableEvcMode_TaskBody struct{
    Req *vsantypes.DisableEvcMode_Task `xml:"urn:vsan DisableEvcMode_Task,omitempty"`
    Res *vsantypes.DisableEvcMode_TaskResponse `xml:"urn:vsan DisableEvcMode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisableEvcMode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DisableEvcMode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisableEvcMode_Task) (*vsantypes.DisableEvcMode_TaskResponse, error) {
  var reqBody, resBody DisableEvcMode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisableFeatureBody struct{
    Req *vsantypes.DisableFeature `xml:"urn:vsan DisableFeature,omitempty"`
    Res *vsantypes.DisableFeatureResponse `xml:"urn:vsan DisableFeatureResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisableFeatureBody) Fault() *soap.Fault { return b.Fault_ }

func DisableFeature(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisableFeature) (*vsantypes.DisableFeatureResponse, error) {
  var reqBody, resBody DisableFeatureBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisableHyperThreadingBody struct{
    Req *vsantypes.DisableHyperThreading `xml:"urn:vsan DisableHyperThreading,omitempty"`
    Res *vsantypes.DisableHyperThreadingResponse `xml:"urn:vsan DisableHyperThreadingResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisableHyperThreadingBody) Fault() *soap.Fault { return b.Fault_ }

func DisableHyperThreading(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisableHyperThreading) (*vsantypes.DisableHyperThreadingResponse, error) {
  var reqBody, resBody DisableHyperThreadingBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisableMultipathPathBody struct{
    Req *vsantypes.DisableMultipathPath `xml:"urn:vsan DisableMultipathPath,omitempty"`
    Res *vsantypes.DisableMultipathPathResponse `xml:"urn:vsan DisableMultipathPathResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisableMultipathPathBody) Fault() *soap.Fault { return b.Fault_ }

func DisableMultipathPath(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisableMultipathPath) (*vsantypes.DisableMultipathPathResponse, error) {
  var reqBody, resBody DisableMultipathPathBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisableRulesetBody struct{
    Req *vsantypes.DisableRuleset `xml:"urn:vsan DisableRuleset,omitempty"`
    Res *vsantypes.DisableRulesetResponse `xml:"urn:vsan DisableRulesetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisableRulesetBody) Fault() *soap.Fault { return b.Fault_ }

func DisableRuleset(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisableRuleset) (*vsantypes.DisableRulesetResponse, error) {
  var reqBody, resBody DisableRulesetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisableSecondaryVM_TaskBody struct{
    Req *vsantypes.DisableSecondaryVM_Task `xml:"urn:vsan DisableSecondaryVM_Task,omitempty"`
    Res *vsantypes.DisableSecondaryVM_TaskResponse `xml:"urn:vsan DisableSecondaryVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisableSecondaryVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DisableSecondaryVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisableSecondaryVM_Task) (*vsantypes.DisableSecondaryVM_TaskResponse, error) {
  var reqBody, resBody DisableSecondaryVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisableSmartCardAuthenticationBody struct{
    Req *vsantypes.DisableSmartCardAuthentication `xml:"urn:vsan DisableSmartCardAuthentication,omitempty"`
    Res *vsantypes.DisableSmartCardAuthenticationResponse `xml:"urn:vsan DisableSmartCardAuthenticationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisableSmartCardAuthenticationBody) Fault() *soap.Fault { return b.Fault_ }

func DisableSmartCardAuthentication(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisableSmartCardAuthentication) (*vsantypes.DisableSmartCardAuthenticationResponse, error) {
  var reqBody, resBody DisableSmartCardAuthenticationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DisconnectHost_TaskBody struct{
    Req *vsantypes.DisconnectHost_Task `xml:"urn:vsan DisconnectHost_Task,omitempty"`
    Res *vsantypes.DisconnectHost_TaskResponse `xml:"urn:vsan DisconnectHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DisconnectHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DisconnectHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DisconnectHost_Task) (*vsantypes.DisconnectHost_TaskResponse, error) {
  var reqBody, resBody DisconnectHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DiscoverFcoeHbasBody struct{
    Req *vsantypes.DiscoverFcoeHbas `xml:"urn:vsan DiscoverFcoeHbas,omitempty"`
    Res *vsantypes.DiscoverFcoeHbasResponse `xml:"urn:vsan DiscoverFcoeHbasResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DiscoverFcoeHbasBody) Fault() *soap.Fault { return b.Fault_ }

func DiscoverFcoeHbas(ctx context.Context, r soap.RoundTripper, req *vsantypes.DiscoverFcoeHbas) (*vsantypes.DiscoverFcoeHbasResponse, error) {
  var reqBody, resBody DiscoverFcoeHbasBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DissociateProfileBody struct{
    Req *vsantypes.DissociateProfile `xml:"urn:vsan DissociateProfile,omitempty"`
    Res *vsantypes.DissociateProfileResponse `xml:"urn:vsan DissociateProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DissociateProfileBody) Fault() *soap.Fault { return b.Fault_ }

func DissociateProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.DissociateProfile) (*vsantypes.DissociateProfileResponse, error) {
  var reqBody, resBody DissociateProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DoesCustomizationSpecExistBody struct{
    Req *vsantypes.DoesCustomizationSpecExist `xml:"urn:vsan DoesCustomizationSpecExist,omitempty"`
    Res *vsantypes.DoesCustomizationSpecExistResponse `xml:"urn:vsan DoesCustomizationSpecExistResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DoesCustomizationSpecExistBody) Fault() *soap.Fault { return b.Fault_ }

func DoesCustomizationSpecExist(ctx context.Context, r soap.RoundTripper, req *vsantypes.DoesCustomizationSpecExist) (*vsantypes.DoesCustomizationSpecExistResponse, error) {
  var reqBody, resBody DoesCustomizationSpecExistBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DuplicateCustomizationSpecBody struct{
    Req *vsantypes.DuplicateCustomizationSpec `xml:"urn:vsan DuplicateCustomizationSpec,omitempty"`
    Res *vsantypes.DuplicateCustomizationSpecResponse `xml:"urn:vsan DuplicateCustomizationSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DuplicateCustomizationSpecBody) Fault() *soap.Fault { return b.Fault_ }

func DuplicateCustomizationSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.DuplicateCustomizationSpec) (*vsantypes.DuplicateCustomizationSpecResponse, error) {
  var reqBody, resBody DuplicateCustomizationSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DvsReconfigureVmVnicNetworkResourcePool_TaskBody struct{
    Req *vsantypes.DvsReconfigureVmVnicNetworkResourcePool_Task `xml:"urn:vsan DvsReconfigureVmVnicNetworkResourcePool_Task,omitempty"`
    Res *vsantypes.DvsReconfigureVmVnicNetworkResourcePool_TaskResponse `xml:"urn:vsan DvsReconfigureVmVnicNetworkResourcePool_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DvsReconfigureVmVnicNetworkResourcePool_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DvsReconfigureVmVnicNetworkResourcePool_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DvsReconfigureVmVnicNetworkResourcePool_Task) (*vsantypes.DvsReconfigureVmVnicNetworkResourcePool_TaskResponse, error) {
  var reqBody, resBody DvsReconfigureVmVnicNetworkResourcePool_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EagerZeroVirtualDisk_TaskBody struct{
    Req *vsantypes.EagerZeroVirtualDisk_Task `xml:"urn:vsan EagerZeroVirtualDisk_Task,omitempty"`
    Res *vsantypes.EagerZeroVirtualDisk_TaskResponse `xml:"urn:vsan EagerZeroVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EagerZeroVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func EagerZeroVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.EagerZeroVirtualDisk_Task) (*vsantypes.EagerZeroVirtualDisk_TaskResponse, error) {
  var reqBody, resBody EagerZeroVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableAlarmActionsBody struct{
    Req *vsantypes.EnableAlarmActions `xml:"urn:vsan EnableAlarmActions,omitempty"`
    Res *vsantypes.EnableAlarmActionsResponse `xml:"urn:vsan EnableAlarmActionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableAlarmActionsBody) Fault() *soap.Fault { return b.Fault_ }

func EnableAlarmActions(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableAlarmActions) (*vsantypes.EnableAlarmActionsResponse, error) {
  var reqBody, resBody EnableAlarmActionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableCryptoBody struct{
    Req *vsantypes.EnableCrypto `xml:"urn:vsan EnableCrypto,omitempty"`
    Res *vsantypes.EnableCryptoResponse `xml:"urn:vsan EnableCryptoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableCryptoBody) Fault() *soap.Fault { return b.Fault_ }

func EnableCrypto(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableCrypto) (*vsantypes.EnableCryptoResponse, error) {
  var reqBody, resBody EnableCryptoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableFeatureBody struct{
    Req *vsantypes.EnableFeature `xml:"urn:vsan EnableFeature,omitempty"`
    Res *vsantypes.EnableFeatureResponse `xml:"urn:vsan EnableFeatureResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableFeatureBody) Fault() *soap.Fault { return b.Fault_ }

func EnableFeature(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableFeature) (*vsantypes.EnableFeatureResponse, error) {
  var reqBody, resBody EnableFeatureBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableHyperThreadingBody struct{
    Req *vsantypes.EnableHyperThreading `xml:"urn:vsan EnableHyperThreading,omitempty"`
    Res *vsantypes.EnableHyperThreadingResponse `xml:"urn:vsan EnableHyperThreadingResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableHyperThreadingBody) Fault() *soap.Fault { return b.Fault_ }

func EnableHyperThreading(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableHyperThreading) (*vsantypes.EnableHyperThreadingResponse, error) {
  var reqBody, resBody EnableHyperThreadingBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableMultipathPathBody struct{
    Req *vsantypes.EnableMultipathPath `xml:"urn:vsan EnableMultipathPath,omitempty"`
    Res *vsantypes.EnableMultipathPathResponse `xml:"urn:vsan EnableMultipathPathResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableMultipathPathBody) Fault() *soap.Fault { return b.Fault_ }

func EnableMultipathPath(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableMultipathPath) (*vsantypes.EnableMultipathPathResponse, error) {
  var reqBody, resBody EnableMultipathPathBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableNetworkResourceManagementBody struct{
    Req *vsantypes.EnableNetworkResourceManagement `xml:"urn:vsan EnableNetworkResourceManagement,omitempty"`
    Res *vsantypes.EnableNetworkResourceManagementResponse `xml:"urn:vsan EnableNetworkResourceManagementResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableNetworkResourceManagementBody) Fault() *soap.Fault { return b.Fault_ }

func EnableNetworkResourceManagement(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableNetworkResourceManagement) (*vsantypes.EnableNetworkResourceManagementResponse, error) {
  var reqBody, resBody EnableNetworkResourceManagementBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableRulesetBody struct{
    Req *vsantypes.EnableRuleset `xml:"urn:vsan EnableRuleset,omitempty"`
    Res *vsantypes.EnableRulesetResponse `xml:"urn:vsan EnableRulesetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableRulesetBody) Fault() *soap.Fault { return b.Fault_ }

func EnableRuleset(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableRuleset) (*vsantypes.EnableRulesetResponse, error) {
  var reqBody, resBody EnableRulesetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableSecondaryVM_TaskBody struct{
    Req *vsantypes.EnableSecondaryVM_Task `xml:"urn:vsan EnableSecondaryVM_Task,omitempty"`
    Res *vsantypes.EnableSecondaryVM_TaskResponse `xml:"urn:vsan EnableSecondaryVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableSecondaryVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func EnableSecondaryVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableSecondaryVM_Task) (*vsantypes.EnableSecondaryVM_TaskResponse, error) {
  var reqBody, resBody EnableSecondaryVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnableSmartCardAuthenticationBody struct{
    Req *vsantypes.EnableSmartCardAuthentication `xml:"urn:vsan EnableSmartCardAuthentication,omitempty"`
    Res *vsantypes.EnableSmartCardAuthenticationResponse `xml:"urn:vsan EnableSmartCardAuthenticationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnableSmartCardAuthenticationBody) Fault() *soap.Fault { return b.Fault_ }

func EnableSmartCardAuthentication(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnableSmartCardAuthentication) (*vsantypes.EnableSmartCardAuthenticationResponse, error) {
  var reqBody, resBody EnableSmartCardAuthenticationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnterLockdownModeBody struct{
    Req *vsantypes.EnterLockdownMode `xml:"urn:vsan EnterLockdownMode,omitempty"`
    Res *vsantypes.EnterLockdownModeResponse `xml:"urn:vsan EnterLockdownModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnterLockdownModeBody) Fault() *soap.Fault { return b.Fault_ }

func EnterLockdownMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnterLockdownMode) (*vsantypes.EnterLockdownModeResponse, error) {
  var reqBody, resBody EnterLockdownModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EnterMaintenanceMode_TaskBody struct{
    Req *vsantypes.EnterMaintenanceMode_Task `xml:"urn:vsan EnterMaintenanceMode_Task,omitempty"`
    Res *vsantypes.EnterMaintenanceMode_TaskResponse `xml:"urn:vsan EnterMaintenanceMode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EnterMaintenanceMode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func EnterMaintenanceMode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.EnterMaintenanceMode_Task) (*vsantypes.EnterMaintenanceMode_TaskResponse, error) {
  var reqBody, resBody EnterMaintenanceMode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EstimateDatabaseSizeBody struct{
    Req *vsantypes.EstimateDatabaseSize `xml:"urn:vsan EstimateDatabaseSize,omitempty"`
    Res *vsantypes.EstimateDatabaseSizeResponse `xml:"urn:vsan EstimateDatabaseSizeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EstimateDatabaseSizeBody) Fault() *soap.Fault { return b.Fault_ }

func EstimateDatabaseSize(ctx context.Context, r soap.RoundTripper, req *vsantypes.EstimateDatabaseSize) (*vsantypes.EstimateDatabaseSizeResponse, error) {
  var reqBody, resBody EstimateDatabaseSizeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EstimateStorageForConsolidateSnapshots_TaskBody struct{
    Req *vsantypes.EstimateStorageForConsolidateSnapshots_Task `xml:"urn:vsan EstimateStorageForConsolidateSnapshots_Task,omitempty"`
    Res *vsantypes.EstimateStorageForConsolidateSnapshots_TaskResponse `xml:"urn:vsan EstimateStorageForConsolidateSnapshots_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EstimateStorageForConsolidateSnapshots_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func EstimateStorageForConsolidateSnapshots_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.EstimateStorageForConsolidateSnapshots_Task) (*vsantypes.EstimateStorageForConsolidateSnapshots_TaskResponse, error) {
  var reqBody, resBody EstimateStorageForConsolidateSnapshots_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EsxAgentHostManagerUpdateConfigBody struct{
    Req *vsantypes.EsxAgentHostManagerUpdateConfig `xml:"urn:vsan EsxAgentHostManagerUpdateConfig,omitempty"`
    Res *vsantypes.EsxAgentHostManagerUpdateConfigResponse `xml:"urn:vsan EsxAgentHostManagerUpdateConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EsxAgentHostManagerUpdateConfigBody) Fault() *soap.Fault { return b.Fault_ }

func EsxAgentHostManagerUpdateConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.EsxAgentHostManagerUpdateConfig) (*vsantypes.EsxAgentHostManagerUpdateConfigResponse, error) {
  var reqBody, resBody EsxAgentHostManagerUpdateConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EvacuateVsanNode_TaskBody struct{
    Req *vsantypes.EvacuateVsanNode_Task `xml:"urn:vsan EvacuateVsanNode_Task,omitempty"`
    Res *vsantypes.EvacuateVsanNode_TaskResponse `xml:"urn:vsan EvacuateVsanNode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EvacuateVsanNode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func EvacuateVsanNode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.EvacuateVsanNode_Task) (*vsantypes.EvacuateVsanNode_TaskResponse, error) {
  var reqBody, resBody EvacuateVsanNode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type EvcManagerBody struct{
    Req *vsantypes.EvcManager `xml:"urn:vsan EvcManager,omitempty"`
    Res *vsantypes.EvcManagerResponse `xml:"urn:vsan EvcManagerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *EvcManagerBody) Fault() *soap.Fault { return b.Fault_ }

func EvcManager(ctx context.Context, r soap.RoundTripper, req *vsantypes.EvcManager) (*vsantypes.EvcManagerResponse, error) {
  var reqBody, resBody EvcManagerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExecuteHostProfileBody struct{
    Req *vsantypes.ExecuteHostProfile `xml:"urn:vsan ExecuteHostProfile,omitempty"`
    Res *vsantypes.ExecuteHostProfileResponse `xml:"urn:vsan ExecuteHostProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExecuteHostProfileBody) Fault() *soap.Fault { return b.Fault_ }

func ExecuteHostProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExecuteHostProfile) (*vsantypes.ExecuteHostProfileResponse, error) {
  var reqBody, resBody ExecuteHostProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExecuteSimpleCommandBody struct{
    Req *vsantypes.ExecuteSimpleCommand `xml:"urn:vsan ExecuteSimpleCommand,omitempty"`
    Res *vsantypes.ExecuteSimpleCommandResponse `xml:"urn:vsan ExecuteSimpleCommandResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExecuteSimpleCommandBody) Fault() *soap.Fault { return b.Fault_ }

func ExecuteSimpleCommand(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExecuteSimpleCommand) (*vsantypes.ExecuteSimpleCommandResponse, error) {
  var reqBody, resBody ExecuteSimpleCommandBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExitLockdownModeBody struct{
    Req *vsantypes.ExitLockdownMode `xml:"urn:vsan ExitLockdownMode,omitempty"`
    Res *vsantypes.ExitLockdownModeResponse `xml:"urn:vsan ExitLockdownModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExitLockdownModeBody) Fault() *soap.Fault { return b.Fault_ }

func ExitLockdownMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExitLockdownMode) (*vsantypes.ExitLockdownModeResponse, error) {
  var reqBody, resBody ExitLockdownModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExitMaintenanceMode_TaskBody struct{
    Req *vsantypes.ExitMaintenanceMode_Task `xml:"urn:vsan ExitMaintenanceMode_Task,omitempty"`
    Res *vsantypes.ExitMaintenanceMode_TaskResponse `xml:"urn:vsan ExitMaintenanceMode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExitMaintenanceMode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ExitMaintenanceMode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExitMaintenanceMode_Task) (*vsantypes.ExitMaintenanceMode_TaskResponse, error) {
  var reqBody, resBody ExitMaintenanceMode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExpandVmfsDatastoreBody struct{
    Req *vsantypes.ExpandVmfsDatastore `xml:"urn:vsan ExpandVmfsDatastore,omitempty"`
    Res *vsantypes.ExpandVmfsDatastoreResponse `xml:"urn:vsan ExpandVmfsDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExpandVmfsDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func ExpandVmfsDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExpandVmfsDatastore) (*vsantypes.ExpandVmfsDatastoreResponse, error) {
  var reqBody, resBody ExpandVmfsDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExpandVmfsExtentBody struct{
    Req *vsantypes.ExpandVmfsExtent `xml:"urn:vsan ExpandVmfsExtent,omitempty"`
    Res *vsantypes.ExpandVmfsExtentResponse `xml:"urn:vsan ExpandVmfsExtentResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExpandVmfsExtentBody) Fault() *soap.Fault { return b.Fault_ }

func ExpandVmfsExtent(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExpandVmfsExtent) (*vsantypes.ExpandVmfsExtentResponse, error) {
  var reqBody, resBody ExpandVmfsExtentBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExportAnswerFile_TaskBody struct{
    Req *vsantypes.ExportAnswerFile_Task `xml:"urn:vsan ExportAnswerFile_Task,omitempty"`
    Res *vsantypes.ExportAnswerFile_TaskResponse `xml:"urn:vsan ExportAnswerFile_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExportAnswerFile_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ExportAnswerFile_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExportAnswerFile_Task) (*vsantypes.ExportAnswerFile_TaskResponse, error) {
  var reqBody, resBody ExportAnswerFile_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExportProfileBody struct{
    Req *vsantypes.ExportProfile `xml:"urn:vsan ExportProfile,omitempty"`
    Res *vsantypes.ExportProfileResponse `xml:"urn:vsan ExportProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExportProfileBody) Fault() *soap.Fault { return b.Fault_ }

func ExportProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExportProfile) (*vsantypes.ExportProfileResponse, error) {
  var reqBody, resBody ExportProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExportSnapshotBody struct{
    Req *vsantypes.ExportSnapshot `xml:"urn:vsan ExportSnapshot,omitempty"`
    Res *vsantypes.ExportSnapshotResponse `xml:"urn:vsan ExportSnapshotResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExportSnapshotBody) Fault() *soap.Fault { return b.Fault_ }

func ExportSnapshot(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExportSnapshot) (*vsantypes.ExportSnapshotResponse, error) {
  var reqBody, resBody ExportSnapshotBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExportVAppBody struct{
    Req *vsantypes.ExportVApp `xml:"urn:vsan ExportVApp,omitempty"`
    Res *vsantypes.ExportVAppResponse `xml:"urn:vsan ExportVAppResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExportVAppBody) Fault() *soap.Fault { return b.Fault_ }

func ExportVApp(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExportVApp) (*vsantypes.ExportVAppResponse, error) {
  var reqBody, resBody ExportVAppBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExportVmBody struct{
    Req *vsantypes.ExportVm `xml:"urn:vsan ExportVm,omitempty"`
    Res *vsantypes.ExportVmResponse `xml:"urn:vsan ExportVmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExportVmBody) Fault() *soap.Fault { return b.Fault_ }

func ExportVm(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExportVm) (*vsantypes.ExportVmResponse, error) {
  var reqBody, resBody ExportVmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExtendDisk_TaskBody struct{
    Req *vsantypes.ExtendDisk_Task `xml:"urn:vsan ExtendDisk_Task,omitempty"`
    Res *vsantypes.ExtendDisk_TaskResponse `xml:"urn:vsan ExtendDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExtendDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ExtendDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExtendDisk_Task) (*vsantypes.ExtendDisk_TaskResponse, error) {
  var reqBody, resBody ExtendDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExtendVffsBody struct{
    Req *vsantypes.ExtendVffs `xml:"urn:vsan ExtendVffs,omitempty"`
    Res *vsantypes.ExtendVffsResponse `xml:"urn:vsan ExtendVffsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExtendVffsBody) Fault() *soap.Fault { return b.Fault_ }

func ExtendVffs(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExtendVffs) (*vsantypes.ExtendVffsResponse, error) {
  var reqBody, resBody ExtendVffsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExtendVirtualDisk_TaskBody struct{
    Req *vsantypes.ExtendVirtualDisk_Task `xml:"urn:vsan ExtendVirtualDisk_Task,omitempty"`
    Res *vsantypes.ExtendVirtualDisk_TaskResponse `xml:"urn:vsan ExtendVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExtendVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ExtendVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExtendVirtualDisk_Task) (*vsantypes.ExtendVirtualDisk_TaskResponse, error) {
  var reqBody, resBody ExtendVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExtendVmfsDatastoreBody struct{
    Req *vsantypes.ExtendVmfsDatastore `xml:"urn:vsan ExtendVmfsDatastore,omitempty"`
    Res *vsantypes.ExtendVmfsDatastoreResponse `xml:"urn:vsan ExtendVmfsDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExtendVmfsDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func ExtendVmfsDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExtendVmfsDatastore) (*vsantypes.ExtendVmfsDatastoreResponse, error) {
  var reqBody, resBody ExtendVmfsDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ExtractOvfEnvironmentBody struct{
    Req *vsantypes.ExtractOvfEnvironment `xml:"urn:vsan ExtractOvfEnvironment,omitempty"`
    Res *vsantypes.ExtractOvfEnvironmentResponse `xml:"urn:vsan ExtractOvfEnvironmentResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ExtractOvfEnvironmentBody) Fault() *soap.Fault { return b.Fault_ }

func ExtractOvfEnvironment(ctx context.Context, r soap.RoundTripper, req *vsantypes.ExtractOvfEnvironment) (*vsantypes.ExtractOvfEnvironmentResponse, error) {
  var reqBody, resBody ExtractOvfEnvironmentBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FetchDVPortKeysBody struct{
    Req *vsantypes.FetchDVPortKeys `xml:"urn:vsan FetchDVPortKeys,omitempty"`
    Res *vsantypes.FetchDVPortKeysResponse `xml:"urn:vsan FetchDVPortKeysResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FetchDVPortKeysBody) Fault() *soap.Fault { return b.Fault_ }

func FetchDVPortKeys(ctx context.Context, r soap.RoundTripper, req *vsantypes.FetchDVPortKeys) (*vsantypes.FetchDVPortKeysResponse, error) {
  var reqBody, resBody FetchDVPortKeysBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FetchDVPortsBody struct{
    Req *vsantypes.FetchDVPorts `xml:"urn:vsan FetchDVPorts,omitempty"`
    Res *vsantypes.FetchDVPortsResponse `xml:"urn:vsan FetchDVPortsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FetchDVPortsBody) Fault() *soap.Fault { return b.Fault_ }

func FetchDVPorts(ctx context.Context, r soap.RoundTripper, req *vsantypes.FetchDVPorts) (*vsantypes.FetchDVPortsResponse, error) {
  var reqBody, resBody FetchDVPortsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FetchSystemEventLogBody struct{
    Req *vsantypes.FetchSystemEventLog `xml:"urn:vsan FetchSystemEventLog,omitempty"`
    Res *vsantypes.FetchSystemEventLogResponse `xml:"urn:vsan FetchSystemEventLogResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FetchSystemEventLogBody) Fault() *soap.Fault { return b.Fault_ }

func FetchSystemEventLog(ctx context.Context, r soap.RoundTripper, req *vsantypes.FetchSystemEventLog) (*vsantypes.FetchSystemEventLogResponse, error) {
  var reqBody, resBody FetchSystemEventLogBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FetchUserPrivilegeOnEntitiesBody struct{
    Req *vsantypes.FetchUserPrivilegeOnEntities `xml:"urn:vsan FetchUserPrivilegeOnEntities,omitempty"`
    Res *vsantypes.FetchUserPrivilegeOnEntitiesResponse `xml:"urn:vsan FetchUserPrivilegeOnEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FetchUserPrivilegeOnEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func FetchUserPrivilegeOnEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.FetchUserPrivilegeOnEntities) (*vsantypes.FetchUserPrivilegeOnEntitiesResponse, error) {
  var reqBody, resBody FetchUserPrivilegeOnEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindAllByDnsNameBody struct{
    Req *vsantypes.FindAllByDnsName `xml:"urn:vsan FindAllByDnsName,omitempty"`
    Res *vsantypes.FindAllByDnsNameResponse `xml:"urn:vsan FindAllByDnsNameResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindAllByDnsNameBody) Fault() *soap.Fault { return b.Fault_ }

func FindAllByDnsName(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindAllByDnsName) (*vsantypes.FindAllByDnsNameResponse, error) {
  var reqBody, resBody FindAllByDnsNameBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindAllByIpBody struct{
    Req *vsantypes.FindAllByIp `xml:"urn:vsan FindAllByIp,omitempty"`
    Res *vsantypes.FindAllByIpResponse `xml:"urn:vsan FindAllByIpResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindAllByIpBody) Fault() *soap.Fault { return b.Fault_ }

func FindAllByIp(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindAllByIp) (*vsantypes.FindAllByIpResponse, error) {
  var reqBody, resBody FindAllByIpBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindAllByUuidBody struct{
    Req *vsantypes.FindAllByUuid `xml:"urn:vsan FindAllByUuid,omitempty"`
    Res *vsantypes.FindAllByUuidResponse `xml:"urn:vsan FindAllByUuidResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindAllByUuidBody) Fault() *soap.Fault { return b.Fault_ }

func FindAllByUuid(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindAllByUuid) (*vsantypes.FindAllByUuidResponse, error) {
  var reqBody, resBody FindAllByUuidBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindAssociatedProfileBody struct{
    Req *vsantypes.FindAssociatedProfile `xml:"urn:vsan FindAssociatedProfile,omitempty"`
    Res *vsantypes.FindAssociatedProfileResponse `xml:"urn:vsan FindAssociatedProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindAssociatedProfileBody) Fault() *soap.Fault { return b.Fault_ }

func FindAssociatedProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindAssociatedProfile) (*vsantypes.FindAssociatedProfileResponse, error) {
  var reqBody, resBody FindAssociatedProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindByDatastorePathBody struct{
    Req *vsantypes.FindByDatastorePath `xml:"urn:vsan FindByDatastorePath,omitempty"`
    Res *vsantypes.FindByDatastorePathResponse `xml:"urn:vsan FindByDatastorePathResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindByDatastorePathBody) Fault() *soap.Fault { return b.Fault_ }

func FindByDatastorePath(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindByDatastorePath) (*vsantypes.FindByDatastorePathResponse, error) {
  var reqBody, resBody FindByDatastorePathBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindByDnsNameBody struct{
    Req *vsantypes.FindByDnsName `xml:"urn:vsan FindByDnsName,omitempty"`
    Res *vsantypes.FindByDnsNameResponse `xml:"urn:vsan FindByDnsNameResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindByDnsNameBody) Fault() *soap.Fault { return b.Fault_ }

func FindByDnsName(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindByDnsName) (*vsantypes.FindByDnsNameResponse, error) {
  var reqBody, resBody FindByDnsNameBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindByInventoryPathBody struct{
    Req *vsantypes.FindByInventoryPath `xml:"urn:vsan FindByInventoryPath,omitempty"`
    Res *vsantypes.FindByInventoryPathResponse `xml:"urn:vsan FindByInventoryPathResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindByInventoryPathBody) Fault() *soap.Fault { return b.Fault_ }

func FindByInventoryPath(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindByInventoryPath) (*vsantypes.FindByInventoryPathResponse, error) {
  var reqBody, resBody FindByInventoryPathBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindByIpBody struct{
    Req *vsantypes.FindByIp `xml:"urn:vsan FindByIp,omitempty"`
    Res *vsantypes.FindByIpResponse `xml:"urn:vsan FindByIpResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindByIpBody) Fault() *soap.Fault { return b.Fault_ }

func FindByIp(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindByIp) (*vsantypes.FindByIpResponse, error) {
  var reqBody, resBody FindByIpBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindByUuidBody struct{
    Req *vsantypes.FindByUuid `xml:"urn:vsan FindByUuid,omitempty"`
    Res *vsantypes.FindByUuidResponse `xml:"urn:vsan FindByUuidResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindByUuidBody) Fault() *soap.Fault { return b.Fault_ }

func FindByUuid(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindByUuid) (*vsantypes.FindByUuidResponse, error) {
  var reqBody, resBody FindByUuidBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindChildBody struct{
    Req *vsantypes.FindChild `xml:"urn:vsan FindChild,omitempty"`
    Res *vsantypes.FindChildResponse `xml:"urn:vsan FindChildResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindChildBody) Fault() *soap.Fault { return b.Fault_ }

func FindChild(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindChild) (*vsantypes.FindChildResponse, error) {
  var reqBody, resBody FindChildBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindExtensionBody struct{
    Req *vsantypes.FindExtension `xml:"urn:vsan FindExtension,omitempty"`
    Res *vsantypes.FindExtensionResponse `xml:"urn:vsan FindExtensionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindExtensionBody) Fault() *soap.Fault { return b.Fault_ }

func FindExtension(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindExtension) (*vsantypes.FindExtensionResponse, error) {
  var reqBody, resBody FindExtensionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FindRulesForVmBody struct{
    Req *vsantypes.FindRulesForVm `xml:"urn:vsan FindRulesForVm,omitempty"`
    Res *vsantypes.FindRulesForVmResponse `xml:"urn:vsan FindRulesForVmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FindRulesForVmBody) Fault() *soap.Fault { return b.Fault_ }

func FindRulesForVm(ctx context.Context, r soap.RoundTripper, req *vsantypes.FindRulesForVm) (*vsantypes.FindRulesForVmResponse, error) {
  var reqBody, resBody FindRulesForVmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FormatVffsBody struct{
    Req *vsantypes.FormatVffs `xml:"urn:vsan FormatVffs,omitempty"`
    Res *vsantypes.FormatVffsResponse `xml:"urn:vsan FormatVffsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FormatVffsBody) Fault() *soap.Fault { return b.Fault_ }

func FormatVffs(ctx context.Context, r soap.RoundTripper, req *vsantypes.FormatVffs) (*vsantypes.FormatVffsResponse, error) {
  var reqBody, resBody FormatVffsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FormatVmfsBody struct{
    Req *vsantypes.FormatVmfs `xml:"urn:vsan FormatVmfs,omitempty"`
    Res *vsantypes.FormatVmfsResponse `xml:"urn:vsan FormatVmfsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FormatVmfsBody) Fault() *soap.Fault { return b.Fault_ }

func FormatVmfs(ctx context.Context, r soap.RoundTripper, req *vsantypes.FormatVmfs) (*vsantypes.FormatVmfsResponse, error) {
  var reqBody, resBody FormatVmfsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateCertificateSigningRequestBody struct{
    Req *vsantypes.GenerateCertificateSigningRequest `xml:"urn:vsan GenerateCertificateSigningRequest,omitempty"`
    Res *vsantypes.GenerateCertificateSigningRequestResponse `xml:"urn:vsan GenerateCertificateSigningRequestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateCertificateSigningRequestBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateCertificateSigningRequest(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateCertificateSigningRequest) (*vsantypes.GenerateCertificateSigningRequestResponse, error) {
  var reqBody, resBody GenerateCertificateSigningRequestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateCertificateSigningRequestByDnBody struct{
    Req *vsantypes.GenerateCertificateSigningRequestByDn `xml:"urn:vsan GenerateCertificateSigningRequestByDn,omitempty"`
    Res *vsantypes.GenerateCertificateSigningRequestByDnResponse `xml:"urn:vsan GenerateCertificateSigningRequestByDnResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateCertificateSigningRequestByDnBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateCertificateSigningRequestByDn(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateCertificateSigningRequestByDn) (*vsantypes.GenerateCertificateSigningRequestByDnResponse, error) {
  var reqBody, resBody GenerateCertificateSigningRequestByDnBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateClientCsrBody struct{
    Req *vsantypes.GenerateClientCsr `xml:"urn:vsan GenerateClientCsr,omitempty"`
    Res *vsantypes.GenerateClientCsrResponse `xml:"urn:vsan GenerateClientCsrResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateClientCsrBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateClientCsr(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateClientCsr) (*vsantypes.GenerateClientCsrResponse, error) {
  var reqBody, resBody GenerateClientCsrBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateConfigTaskListBody struct{
    Req *vsantypes.GenerateConfigTaskList `xml:"urn:vsan GenerateConfigTaskList,omitempty"`
    Res *vsantypes.GenerateConfigTaskListResponse `xml:"urn:vsan GenerateConfigTaskListResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateConfigTaskListBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateConfigTaskList(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateConfigTaskList) (*vsantypes.GenerateConfigTaskListResponse, error) {
  var reqBody, resBody GenerateConfigTaskListBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateHostConfigTaskSpec_TaskBody struct{
    Req *vsantypes.GenerateHostConfigTaskSpec_Task `xml:"urn:vsan GenerateHostConfigTaskSpec_Task,omitempty"`
    Res *vsantypes.GenerateHostConfigTaskSpec_TaskResponse `xml:"urn:vsan GenerateHostConfigTaskSpec_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateHostConfigTaskSpec_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateHostConfigTaskSpec_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateHostConfigTaskSpec_Task) (*vsantypes.GenerateHostConfigTaskSpec_TaskResponse, error) {
  var reqBody, resBody GenerateHostConfigTaskSpec_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateHostProfileTaskList_TaskBody struct{
    Req *vsantypes.GenerateHostProfileTaskList_Task `xml:"urn:vsan GenerateHostProfileTaskList_Task,omitempty"`
    Res *vsantypes.GenerateHostProfileTaskList_TaskResponse `xml:"urn:vsan GenerateHostProfileTaskList_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateHostProfileTaskList_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateHostProfileTaskList_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateHostProfileTaskList_Task) (*vsantypes.GenerateHostProfileTaskList_TaskResponse, error) {
  var reqBody, resBody GenerateHostProfileTaskList_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateKeyBody struct{
    Req *vsantypes.GenerateKey `xml:"urn:vsan GenerateKey,omitempty"`
    Res *vsantypes.GenerateKeyResponse `xml:"urn:vsan GenerateKeyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateKeyBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateKey(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateKey) (*vsantypes.GenerateKeyResponse, error) {
  var reqBody, resBody GenerateKeyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateLogBundles_TaskBody struct{
    Req *vsantypes.GenerateLogBundles_Task `xml:"urn:vsan GenerateLogBundles_Task,omitempty"`
    Res *vsantypes.GenerateLogBundles_TaskResponse `xml:"urn:vsan GenerateLogBundles_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateLogBundles_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateLogBundles_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateLogBundles_Task) (*vsantypes.GenerateLogBundles_TaskResponse, error) {
  var reqBody, resBody GenerateLogBundles_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GenerateSelfSignedClientCertBody struct{
    Req *vsantypes.GenerateSelfSignedClientCert `xml:"urn:vsan GenerateSelfSignedClientCert,omitempty"`
    Res *vsantypes.GenerateSelfSignedClientCertResponse `xml:"urn:vsan GenerateSelfSignedClientCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GenerateSelfSignedClientCertBody) Fault() *soap.Fault { return b.Fault_ }

func GenerateSelfSignedClientCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.GenerateSelfSignedClientCert) (*vsantypes.GenerateSelfSignedClientCertResponse, error) {
  var reqBody, resBody GenerateSelfSignedClientCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetAlarmBody struct{
    Req *vsantypes.GetAlarm `xml:"urn:vsan GetAlarm,omitempty"`
    Res *vsantypes.GetAlarmResponse `xml:"urn:vsan GetAlarmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetAlarmBody) Fault() *soap.Fault { return b.Fault_ }

func GetAlarm(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetAlarm) (*vsantypes.GetAlarmResponse, error) {
  var reqBody, resBody GetAlarmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetAlarmStateBody struct{
    Req *vsantypes.GetAlarmState `xml:"urn:vsan GetAlarmState,omitempty"`
    Res *vsantypes.GetAlarmStateResponse `xml:"urn:vsan GetAlarmStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetAlarmStateBody) Fault() *soap.Fault { return b.Fault_ }

func GetAlarmState(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetAlarmState) (*vsantypes.GetAlarmStateResponse, error) {
  var reqBody, resBody GetAlarmStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetCustomizationSpecBody struct{
    Req *vsantypes.GetCustomizationSpec `xml:"urn:vsan GetCustomizationSpec,omitempty"`
    Res *vsantypes.GetCustomizationSpecResponse `xml:"urn:vsan GetCustomizationSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetCustomizationSpecBody) Fault() *soap.Fault { return b.Fault_ }

func GetCustomizationSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetCustomizationSpec) (*vsantypes.GetCustomizationSpecResponse, error) {
  var reqBody, resBody GetCustomizationSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetPublicKeyBody struct{
    Req *vsantypes.GetPublicKey `xml:"urn:vsan GetPublicKey,omitempty"`
    Res *vsantypes.GetPublicKeyResponse `xml:"urn:vsan GetPublicKeyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetPublicKeyBody) Fault() *soap.Fault { return b.Fault_ }

func GetPublicKey(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetPublicKey) (*vsantypes.GetPublicKeyResponse, error) {
  var reqBody, resBody GetPublicKeyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetResourceUsageBody struct{
    Req *vsantypes.GetResourceUsage `xml:"urn:vsan GetResourceUsage,omitempty"`
    Res *vsantypes.GetResourceUsageResponse `xml:"urn:vsan GetResourceUsageResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetResourceUsageBody) Fault() *soap.Fault { return b.Fault_ }

func GetResourceUsage(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetResourceUsage) (*vsantypes.GetResourceUsageResponse, error) {
  var reqBody, resBody GetResourceUsageBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetVchaClusterHealthBody struct{
    Req *vsantypes.GetVchaClusterHealth `xml:"urn:vsan GetVchaClusterHealth,omitempty"`
    Res *vsantypes.GetVchaClusterHealthResponse `xml:"urn:vsan GetVchaClusterHealthResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetVchaClusterHealthBody) Fault() *soap.Fault { return b.Fault_ }

func GetVchaClusterHealth(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetVchaClusterHealth) (*vsantypes.GetVchaClusterHealthResponse, error) {
  var reqBody, resBody GetVchaClusterHealthBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetVsanObjExtAttrsBody struct{
    Req *vsantypes.GetVsanObjExtAttrs `xml:"urn:vsan GetVsanObjExtAttrs,omitempty"`
    Res *vsantypes.GetVsanObjExtAttrsResponse `xml:"urn:vsan GetVsanObjExtAttrsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetVsanObjExtAttrsBody) Fault() *soap.Fault { return b.Fault_ }

func GetVsanObjExtAttrs(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetVsanObjExtAttrs) (*vsantypes.GetVsanObjExtAttrsResponse, error) {
  var reqBody, resBody GetVsanObjExtAttrsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HasMonitoredEntityBody struct{
    Req *vsantypes.HasMonitoredEntity `xml:"urn:vsan HasMonitoredEntity,omitempty"`
    Res *vsantypes.HasMonitoredEntityResponse `xml:"urn:vsan HasMonitoredEntityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HasMonitoredEntityBody) Fault() *soap.Fault { return b.Fault_ }

func HasMonitoredEntity(ctx context.Context, r soap.RoundTripper, req *vsantypes.HasMonitoredEntity) (*vsantypes.HasMonitoredEntityResponse, error) {
  var reqBody, resBody HasMonitoredEntityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HasPrivilegeOnEntitiesBody struct{
    Req *vsantypes.HasPrivilegeOnEntities `xml:"urn:vsan HasPrivilegeOnEntities,omitempty"`
    Res *vsantypes.HasPrivilegeOnEntitiesResponse `xml:"urn:vsan HasPrivilegeOnEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HasPrivilegeOnEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func HasPrivilegeOnEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.HasPrivilegeOnEntities) (*vsantypes.HasPrivilegeOnEntitiesResponse, error) {
  var reqBody, resBody HasPrivilegeOnEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HasPrivilegeOnEntityBody struct{
    Req *vsantypes.HasPrivilegeOnEntity `xml:"urn:vsan HasPrivilegeOnEntity,omitempty"`
    Res *vsantypes.HasPrivilegeOnEntityResponse `xml:"urn:vsan HasPrivilegeOnEntityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HasPrivilegeOnEntityBody) Fault() *soap.Fault { return b.Fault_ }

func HasPrivilegeOnEntity(ctx context.Context, r soap.RoundTripper, req *vsantypes.HasPrivilegeOnEntity) (*vsantypes.HasPrivilegeOnEntityResponse, error) {
  var reqBody, resBody HasPrivilegeOnEntityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HasProviderBody struct{
    Req *vsantypes.HasProvider `xml:"urn:vsan HasProvider,omitempty"`
    Res *vsantypes.HasProviderResponse `xml:"urn:vsan HasProviderResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HasProviderBody) Fault() *soap.Fault { return b.Fault_ }

func HasProvider(ctx context.Context, r soap.RoundTripper, req *vsantypes.HasProvider) (*vsantypes.HasProviderResponse, error) {
  var reqBody, resBody HasProviderBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HasUserPrivilegeOnEntitiesBody struct{
    Req *vsantypes.HasUserPrivilegeOnEntities `xml:"urn:vsan HasUserPrivilegeOnEntities,omitempty"`
    Res *vsantypes.HasUserPrivilegeOnEntitiesResponse `xml:"urn:vsan HasUserPrivilegeOnEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HasUserPrivilegeOnEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func HasUserPrivilegeOnEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.HasUserPrivilegeOnEntities) (*vsantypes.HasUserPrivilegeOnEntitiesResponse, error) {
  var reqBody, resBody HasUserPrivilegeOnEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostCloneVStorageObject_TaskBody struct{
    Req *vsantypes.HostCloneVStorageObject_Task `xml:"urn:vsan HostCloneVStorageObject_Task,omitempty"`
    Res *vsantypes.HostCloneVStorageObject_TaskResponse `xml:"urn:vsan HostCloneVStorageObject_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostCloneVStorageObject_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func HostCloneVStorageObject_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostCloneVStorageObject_Task) (*vsantypes.HostCloneVStorageObject_TaskResponse, error) {
  var reqBody, resBody HostCloneVStorageObject_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostConfigVFlashCacheBody struct{
    Req *vsantypes.HostConfigVFlashCache `xml:"urn:vsan HostConfigVFlashCache,omitempty"`
    Res *vsantypes.HostConfigVFlashCacheResponse `xml:"urn:vsan HostConfigVFlashCacheResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostConfigVFlashCacheBody) Fault() *soap.Fault { return b.Fault_ }

func HostConfigVFlashCache(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostConfigVFlashCache) (*vsantypes.HostConfigVFlashCacheResponse, error) {
  var reqBody, resBody HostConfigVFlashCacheBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostConfigureVFlashResourceBody struct{
    Req *vsantypes.HostConfigureVFlashResource `xml:"urn:vsan HostConfigureVFlashResource,omitempty"`
    Res *vsantypes.HostConfigureVFlashResourceResponse `xml:"urn:vsan HostConfigureVFlashResourceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostConfigureVFlashResourceBody) Fault() *soap.Fault { return b.Fault_ }

func HostConfigureVFlashResource(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostConfigureVFlashResource) (*vsantypes.HostConfigureVFlashResourceResponse, error) {
  var reqBody, resBody HostConfigureVFlashResourceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostCreateDisk_TaskBody struct{
    Req *vsantypes.HostCreateDisk_Task `xml:"urn:vsan HostCreateDisk_Task,omitempty"`
    Res *vsantypes.HostCreateDisk_TaskResponse `xml:"urn:vsan HostCreateDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostCreateDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func HostCreateDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostCreateDisk_Task) (*vsantypes.HostCreateDisk_TaskResponse, error) {
  var reqBody, resBody HostCreateDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostDeleteVStorageObject_TaskBody struct{
    Req *vsantypes.HostDeleteVStorageObject_Task `xml:"urn:vsan HostDeleteVStorageObject_Task,omitempty"`
    Res *vsantypes.HostDeleteVStorageObject_TaskResponse `xml:"urn:vsan HostDeleteVStorageObject_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostDeleteVStorageObject_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func HostDeleteVStorageObject_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostDeleteVStorageObject_Task) (*vsantypes.HostDeleteVStorageObject_TaskResponse, error) {
  var reqBody, resBody HostDeleteVStorageObject_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostExtendDisk_TaskBody struct{
    Req *vsantypes.HostExtendDisk_Task `xml:"urn:vsan HostExtendDisk_Task,omitempty"`
    Res *vsantypes.HostExtendDisk_TaskResponse `xml:"urn:vsan HostExtendDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostExtendDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func HostExtendDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostExtendDisk_Task) (*vsantypes.HostExtendDisk_TaskResponse, error) {
  var reqBody, resBody HostExtendDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostGetVFlashModuleDefaultConfigBody struct{
    Req *vsantypes.HostGetVFlashModuleDefaultConfig `xml:"urn:vsan HostGetVFlashModuleDefaultConfig,omitempty"`
    Res *vsantypes.HostGetVFlashModuleDefaultConfigResponse `xml:"urn:vsan HostGetVFlashModuleDefaultConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostGetVFlashModuleDefaultConfigBody) Fault() *soap.Fault { return b.Fault_ }

func HostGetVFlashModuleDefaultConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostGetVFlashModuleDefaultConfig) (*vsantypes.HostGetVFlashModuleDefaultConfigResponse, error) {
  var reqBody, resBody HostGetVFlashModuleDefaultConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostImageConfigGetAcceptanceBody struct{
    Req *vsantypes.HostImageConfigGetAcceptance `xml:"urn:vsan HostImageConfigGetAcceptance,omitempty"`
    Res *vsantypes.HostImageConfigGetAcceptanceResponse `xml:"urn:vsan HostImageConfigGetAcceptanceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostImageConfigGetAcceptanceBody) Fault() *soap.Fault { return b.Fault_ }

func HostImageConfigGetAcceptance(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostImageConfigGetAcceptance) (*vsantypes.HostImageConfigGetAcceptanceResponse, error) {
  var reqBody, resBody HostImageConfigGetAcceptanceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostImageConfigGetProfileBody struct{
    Req *vsantypes.HostImageConfigGetProfile `xml:"urn:vsan HostImageConfigGetProfile,omitempty"`
    Res *vsantypes.HostImageConfigGetProfileResponse `xml:"urn:vsan HostImageConfigGetProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostImageConfigGetProfileBody) Fault() *soap.Fault { return b.Fault_ }

func HostImageConfigGetProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostImageConfigGetProfile) (*vsantypes.HostImageConfigGetProfileResponse, error) {
  var reqBody, resBody HostImageConfigGetProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostInflateDisk_TaskBody struct{
    Req *vsantypes.HostInflateDisk_Task `xml:"urn:vsan HostInflateDisk_Task,omitempty"`
    Res *vsantypes.HostInflateDisk_TaskResponse `xml:"urn:vsan HostInflateDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostInflateDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func HostInflateDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostInflateDisk_Task) (*vsantypes.HostInflateDisk_TaskResponse, error) {
  var reqBody, resBody HostInflateDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostListVStorageObjectBody struct{
    Req *vsantypes.HostListVStorageObject `xml:"urn:vsan HostListVStorageObject,omitempty"`
    Res *vsantypes.HostListVStorageObjectResponse `xml:"urn:vsan HostListVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostListVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func HostListVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostListVStorageObject) (*vsantypes.HostListVStorageObjectResponse, error) {
  var reqBody, resBody HostListVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostReconcileDatastoreInventory_TaskBody struct{
    Req *vsantypes.HostReconcileDatastoreInventory_Task `xml:"urn:vsan HostReconcileDatastoreInventory_Task,omitempty"`
    Res *vsantypes.HostReconcileDatastoreInventory_TaskResponse `xml:"urn:vsan HostReconcileDatastoreInventory_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostReconcileDatastoreInventory_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func HostReconcileDatastoreInventory_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostReconcileDatastoreInventory_Task) (*vsantypes.HostReconcileDatastoreInventory_TaskResponse, error) {
  var reqBody, resBody HostReconcileDatastoreInventory_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostRegisterDiskBody struct{
    Req *vsantypes.HostRegisterDisk `xml:"urn:vsan HostRegisterDisk,omitempty"`
    Res *vsantypes.HostRegisterDiskResponse `xml:"urn:vsan HostRegisterDiskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostRegisterDiskBody) Fault() *soap.Fault { return b.Fault_ }

func HostRegisterDisk(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostRegisterDisk) (*vsantypes.HostRegisterDiskResponse, error) {
  var reqBody, resBody HostRegisterDiskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostRelocateVStorageObject_TaskBody struct{
    Req *vsantypes.HostRelocateVStorageObject_Task `xml:"urn:vsan HostRelocateVStorageObject_Task,omitempty"`
    Res *vsantypes.HostRelocateVStorageObject_TaskResponse `xml:"urn:vsan HostRelocateVStorageObject_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostRelocateVStorageObject_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func HostRelocateVStorageObject_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostRelocateVStorageObject_Task) (*vsantypes.HostRelocateVStorageObject_TaskResponse, error) {
  var reqBody, resBody HostRelocateVStorageObject_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostRemoveVFlashResourceBody struct{
    Req *vsantypes.HostRemoveVFlashResource `xml:"urn:vsan HostRemoveVFlashResource,omitempty"`
    Res *vsantypes.HostRemoveVFlashResourceResponse `xml:"urn:vsan HostRemoveVFlashResourceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostRemoveVFlashResourceBody) Fault() *soap.Fault { return b.Fault_ }

func HostRemoveVFlashResource(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostRemoveVFlashResource) (*vsantypes.HostRemoveVFlashResourceResponse, error) {
  var reqBody, resBody HostRemoveVFlashResourceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostRenameVStorageObjectBody struct{
    Req *vsantypes.HostRenameVStorageObject `xml:"urn:vsan HostRenameVStorageObject,omitempty"`
    Res *vsantypes.HostRenameVStorageObjectResponse `xml:"urn:vsan HostRenameVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostRenameVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func HostRenameVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostRenameVStorageObject) (*vsantypes.HostRenameVStorageObjectResponse, error) {
  var reqBody, resBody HostRenameVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostRetrieveVStorageObjectBody struct{
    Req *vsantypes.HostRetrieveVStorageObject `xml:"urn:vsan HostRetrieveVStorageObject,omitempty"`
    Res *vsantypes.HostRetrieveVStorageObjectResponse `xml:"urn:vsan HostRetrieveVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostRetrieveVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func HostRetrieveVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostRetrieveVStorageObject) (*vsantypes.HostRetrieveVStorageObjectResponse, error) {
  var reqBody, resBody HostRetrieveVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostRetrieveVStorageObjectStateBody struct{
    Req *vsantypes.HostRetrieveVStorageObjectState `xml:"urn:vsan HostRetrieveVStorageObjectState,omitempty"`
    Res *vsantypes.HostRetrieveVStorageObjectStateResponse `xml:"urn:vsan HostRetrieveVStorageObjectStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostRetrieveVStorageObjectStateBody) Fault() *soap.Fault { return b.Fault_ }

func HostRetrieveVStorageObjectState(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostRetrieveVStorageObjectState) (*vsantypes.HostRetrieveVStorageObjectStateResponse, error) {
  var reqBody, resBody HostRetrieveVStorageObjectStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostScheduleReconcileDatastoreInventoryBody struct{
    Req *vsantypes.HostScheduleReconcileDatastoreInventory `xml:"urn:vsan HostScheduleReconcileDatastoreInventory,omitempty"`
    Res *vsantypes.HostScheduleReconcileDatastoreInventoryResponse `xml:"urn:vsan HostScheduleReconcileDatastoreInventoryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostScheduleReconcileDatastoreInventoryBody) Fault() *soap.Fault { return b.Fault_ }

func HostScheduleReconcileDatastoreInventory(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostScheduleReconcileDatastoreInventory) (*vsantypes.HostScheduleReconcileDatastoreInventoryResponse, error) {
  var reqBody, resBody HostScheduleReconcileDatastoreInventoryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HostSpecGetUpdatedHostsBody struct{
    Req *vsantypes.HostSpecGetUpdatedHosts `xml:"urn:vsan HostSpecGetUpdatedHosts,omitempty"`
    Res *vsantypes.HostSpecGetUpdatedHostsResponse `xml:"urn:vsan HostSpecGetUpdatedHostsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HostSpecGetUpdatedHostsBody) Fault() *soap.Fault { return b.Fault_ }

func HostSpecGetUpdatedHosts(ctx context.Context, r soap.RoundTripper, req *vsantypes.HostSpecGetUpdatedHosts) (*vsantypes.HostSpecGetUpdatedHostsResponse, error) {
  var reqBody, resBody HostSpecGetUpdatedHostsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HttpNfcLeaseAbortBody struct{
    Req *vsantypes.HttpNfcLeaseAbort `xml:"urn:vsan HttpNfcLeaseAbort,omitempty"`
    Res *vsantypes.HttpNfcLeaseAbortResponse `xml:"urn:vsan HttpNfcLeaseAbortResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HttpNfcLeaseAbortBody) Fault() *soap.Fault { return b.Fault_ }

func HttpNfcLeaseAbort(ctx context.Context, r soap.RoundTripper, req *vsantypes.HttpNfcLeaseAbort) (*vsantypes.HttpNfcLeaseAbortResponse, error) {
  var reqBody, resBody HttpNfcLeaseAbortBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HttpNfcLeaseCompleteBody struct{
    Req *vsantypes.HttpNfcLeaseComplete `xml:"urn:vsan HttpNfcLeaseComplete,omitempty"`
    Res *vsantypes.HttpNfcLeaseCompleteResponse `xml:"urn:vsan HttpNfcLeaseCompleteResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HttpNfcLeaseCompleteBody) Fault() *soap.Fault { return b.Fault_ }

func HttpNfcLeaseComplete(ctx context.Context, r soap.RoundTripper, req *vsantypes.HttpNfcLeaseComplete) (*vsantypes.HttpNfcLeaseCompleteResponse, error) {
  var reqBody, resBody HttpNfcLeaseCompleteBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HttpNfcLeaseGetManifestBody struct{
    Req *vsantypes.HttpNfcLeaseGetManifest `xml:"urn:vsan HttpNfcLeaseGetManifest,omitempty"`
    Res *vsantypes.HttpNfcLeaseGetManifestResponse `xml:"urn:vsan HttpNfcLeaseGetManifestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HttpNfcLeaseGetManifestBody) Fault() *soap.Fault { return b.Fault_ }

func HttpNfcLeaseGetManifest(ctx context.Context, r soap.RoundTripper, req *vsantypes.HttpNfcLeaseGetManifest) (*vsantypes.HttpNfcLeaseGetManifestResponse, error) {
  var reqBody, resBody HttpNfcLeaseGetManifestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type HttpNfcLeaseProgressBody struct{
    Req *vsantypes.HttpNfcLeaseProgress `xml:"urn:vsan HttpNfcLeaseProgress,omitempty"`
    Res *vsantypes.HttpNfcLeaseProgressResponse `xml:"urn:vsan HttpNfcLeaseProgressResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *HttpNfcLeaseProgressBody) Fault() *soap.Fault { return b.Fault_ }

func HttpNfcLeaseProgress(ctx context.Context, r soap.RoundTripper, req *vsantypes.HttpNfcLeaseProgress) (*vsantypes.HttpNfcLeaseProgressResponse, error) {
  var reqBody, resBody HttpNfcLeaseProgressBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ImpersonateUserBody struct{
    Req *vsantypes.ImpersonateUser `xml:"urn:vsan ImpersonateUser,omitempty"`
    Res *vsantypes.ImpersonateUserResponse `xml:"urn:vsan ImpersonateUserResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ImpersonateUserBody) Fault() *soap.Fault { return b.Fault_ }

func ImpersonateUser(ctx context.Context, r soap.RoundTripper, req *vsantypes.ImpersonateUser) (*vsantypes.ImpersonateUserResponse, error) {
  var reqBody, resBody ImpersonateUserBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ImportCertificateForCAM_TaskBody struct{
    Req *vsantypes.ImportCertificateForCAM_Task `xml:"urn:vsan ImportCertificateForCAM_Task,omitempty"`
    Res *vsantypes.ImportCertificateForCAM_TaskResponse `xml:"urn:vsan ImportCertificateForCAM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ImportCertificateForCAM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ImportCertificateForCAM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ImportCertificateForCAM_Task) (*vsantypes.ImportCertificateForCAM_TaskResponse, error) {
  var reqBody, resBody ImportCertificateForCAM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ImportUnmanagedSnapshotBody struct{
    Req *vsantypes.ImportUnmanagedSnapshot `xml:"urn:vsan ImportUnmanagedSnapshot,omitempty"`
    Res *vsantypes.ImportUnmanagedSnapshotResponse `xml:"urn:vsan ImportUnmanagedSnapshotResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ImportUnmanagedSnapshotBody) Fault() *soap.Fault { return b.Fault_ }

func ImportUnmanagedSnapshot(ctx context.Context, r soap.RoundTripper, req *vsantypes.ImportUnmanagedSnapshot) (*vsantypes.ImportUnmanagedSnapshotResponse, error) {
  var reqBody, resBody ImportUnmanagedSnapshotBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ImportVAppBody struct{
    Req *vsantypes.ImportVApp `xml:"urn:vsan ImportVApp,omitempty"`
    Res *vsantypes.ImportVAppResponse `xml:"urn:vsan ImportVAppResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ImportVAppBody) Fault() *soap.Fault { return b.Fault_ }

func ImportVApp(ctx context.Context, r soap.RoundTripper, req *vsantypes.ImportVApp) (*vsantypes.ImportVAppResponse, error) {
  var reqBody, resBody ImportVAppBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InflateDisk_TaskBody struct{
    Req *vsantypes.InflateDisk_Task `xml:"urn:vsan InflateDisk_Task,omitempty"`
    Res *vsantypes.InflateDisk_TaskResponse `xml:"urn:vsan InflateDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InflateDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func InflateDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.InflateDisk_Task) (*vsantypes.InflateDisk_TaskResponse, error) {
  var reqBody, resBody InflateDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InflateVirtualDisk_TaskBody struct{
    Req *vsantypes.InflateVirtualDisk_Task `xml:"urn:vsan InflateVirtualDisk_Task,omitempty"`
    Res *vsantypes.InflateVirtualDisk_TaskResponse `xml:"urn:vsan InflateVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InflateVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func InflateVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.InflateVirtualDisk_Task) (*vsantypes.InflateVirtualDisk_TaskResponse, error) {
  var reqBody, resBody InflateVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InitializeDiskMappingsBody struct{
    Req *vsantypes.InitializeDiskMappings `xml:"urn:vsan InitializeDiskMappings,omitempty"`
    Res *vsantypes.InitializeDiskMappingsResponse `xml:"urn:vsan InitializeDiskMappingsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InitializeDiskMappingsBody) Fault() *soap.Fault { return b.Fault_ }

func InitializeDiskMappings(ctx context.Context, r soap.RoundTripper, req *vsantypes.InitializeDiskMappings) (*vsantypes.InitializeDiskMappingsResponse, error) {
  var reqBody, resBody InitializeDiskMappingsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InitializeDisks_TaskBody struct{
    Req *vsantypes.InitializeDisks_Task `xml:"urn:vsan InitializeDisks_Task,omitempty"`
    Res *vsantypes.InitializeDisks_TaskResponse `xml:"urn:vsan InitializeDisks_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InitializeDisks_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func InitializeDisks_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.InitializeDisks_Task) (*vsantypes.InitializeDisks_TaskResponse, error) {
  var reqBody, resBody InitializeDisks_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InitiateFileTransferFromGuestBody struct{
    Req *vsantypes.InitiateFileTransferFromGuest `xml:"urn:vsan InitiateFileTransferFromGuest,omitempty"`
    Res *vsantypes.InitiateFileTransferFromGuestResponse `xml:"urn:vsan InitiateFileTransferFromGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InitiateFileTransferFromGuestBody) Fault() *soap.Fault { return b.Fault_ }

func InitiateFileTransferFromGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.InitiateFileTransferFromGuest) (*vsantypes.InitiateFileTransferFromGuestResponse, error) {
  var reqBody, resBody InitiateFileTransferFromGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InitiateFileTransferToGuestBody struct{
    Req *vsantypes.InitiateFileTransferToGuest `xml:"urn:vsan InitiateFileTransferToGuest,omitempty"`
    Res *vsantypes.InitiateFileTransferToGuestResponse `xml:"urn:vsan InitiateFileTransferToGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InitiateFileTransferToGuestBody) Fault() *soap.Fault { return b.Fault_ }

func InitiateFileTransferToGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.InitiateFileTransferToGuest) (*vsantypes.InitiateFileTransferToGuestResponse, error) {
  var reqBody, resBody InitiateFileTransferToGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InstallHostPatchV2_TaskBody struct{
    Req *vsantypes.InstallHostPatchV2_Task `xml:"urn:vsan InstallHostPatchV2_Task,omitempty"`
    Res *vsantypes.InstallHostPatchV2_TaskResponse `xml:"urn:vsan InstallHostPatchV2_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InstallHostPatchV2_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func InstallHostPatchV2_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.InstallHostPatchV2_Task) (*vsantypes.InstallHostPatchV2_TaskResponse, error) {
  var reqBody, resBody InstallHostPatchV2_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InstallHostPatch_TaskBody struct{
    Req *vsantypes.InstallHostPatch_Task `xml:"urn:vsan InstallHostPatch_Task,omitempty"`
    Res *vsantypes.InstallHostPatch_TaskResponse `xml:"urn:vsan InstallHostPatch_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InstallHostPatch_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func InstallHostPatch_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.InstallHostPatch_Task) (*vsantypes.InstallHostPatch_TaskResponse, error) {
  var reqBody, resBody InstallHostPatch_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InstallIoFilter_TaskBody struct{
    Req *vsantypes.InstallIoFilter_Task `xml:"urn:vsan InstallIoFilter_Task,omitempty"`
    Res *vsantypes.InstallIoFilter_TaskResponse `xml:"urn:vsan InstallIoFilter_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InstallIoFilter_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func InstallIoFilter_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.InstallIoFilter_Task) (*vsantypes.InstallIoFilter_TaskResponse, error) {
  var reqBody, resBody InstallIoFilter_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InstallServerCertificateBody struct{
    Req *vsantypes.InstallServerCertificate `xml:"urn:vsan InstallServerCertificate,omitempty"`
    Res *vsantypes.InstallServerCertificateResponse `xml:"urn:vsan InstallServerCertificateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InstallServerCertificateBody) Fault() *soap.Fault { return b.Fault_ }

func InstallServerCertificate(ctx context.Context, r soap.RoundTripper, req *vsantypes.InstallServerCertificate) (*vsantypes.InstallServerCertificateResponse, error) {
  var reqBody, resBody InstallServerCertificateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InstallSmartCardTrustAnchorBody struct{
    Req *vsantypes.InstallSmartCardTrustAnchor `xml:"urn:vsan InstallSmartCardTrustAnchor,omitempty"`
    Res *vsantypes.InstallSmartCardTrustAnchorResponse `xml:"urn:vsan InstallSmartCardTrustAnchorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InstallSmartCardTrustAnchorBody) Fault() *soap.Fault { return b.Fault_ }

func InstallSmartCardTrustAnchor(ctx context.Context, r soap.RoundTripper, req *vsantypes.InstallSmartCardTrustAnchor) (*vsantypes.InstallSmartCardTrustAnchorResponse, error) {
  var reqBody, resBody InstallSmartCardTrustAnchorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type IsSharedGraphicsActiveBody struct{
    Req *vsantypes.IsSharedGraphicsActive `xml:"urn:vsan IsSharedGraphicsActive,omitempty"`
    Res *vsantypes.IsSharedGraphicsActiveResponse `xml:"urn:vsan IsSharedGraphicsActiveResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *IsSharedGraphicsActiveBody) Fault() *soap.Fault { return b.Fault_ }

func IsSharedGraphicsActive(ctx context.Context, r soap.RoundTripper, req *vsantypes.IsSharedGraphicsActive) (*vsantypes.IsSharedGraphicsActiveResponse, error) {
  var reqBody, resBody IsSharedGraphicsActiveBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type JoinDomainWithCAM_TaskBody struct{
    Req *vsantypes.JoinDomainWithCAM_Task `xml:"urn:vsan JoinDomainWithCAM_Task,omitempty"`
    Res *vsantypes.JoinDomainWithCAM_TaskResponse `xml:"urn:vsan JoinDomainWithCAM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *JoinDomainWithCAM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func JoinDomainWithCAM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.JoinDomainWithCAM_Task) (*vsantypes.JoinDomainWithCAM_TaskResponse, error) {
  var reqBody, resBody JoinDomainWithCAM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type JoinDomain_TaskBody struct{
    Req *vsantypes.JoinDomain_Task `xml:"urn:vsan JoinDomain_Task,omitempty"`
    Res *vsantypes.JoinDomain_TaskResponse `xml:"urn:vsan JoinDomain_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *JoinDomain_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func JoinDomain_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.JoinDomain_Task) (*vsantypes.JoinDomain_TaskResponse, error) {
  var reqBody, resBody JoinDomain_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LeaveCurrentDomain_TaskBody struct{
    Req *vsantypes.LeaveCurrentDomain_Task `xml:"urn:vsan LeaveCurrentDomain_Task,omitempty"`
    Res *vsantypes.LeaveCurrentDomain_TaskResponse `xml:"urn:vsan LeaveCurrentDomain_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LeaveCurrentDomain_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func LeaveCurrentDomain_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.LeaveCurrentDomain_Task) (*vsantypes.LeaveCurrentDomain_TaskResponse, error) {
  var reqBody, resBody LeaveCurrentDomain_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListCACertificateRevocationListsBody struct{
    Req *vsantypes.ListCACertificateRevocationLists `xml:"urn:vsan ListCACertificateRevocationLists,omitempty"`
    Res *vsantypes.ListCACertificateRevocationListsResponse `xml:"urn:vsan ListCACertificateRevocationListsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListCACertificateRevocationListsBody) Fault() *soap.Fault { return b.Fault_ }

func ListCACertificateRevocationLists(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListCACertificateRevocationLists) (*vsantypes.ListCACertificateRevocationListsResponse, error) {
  var reqBody, resBody ListCACertificateRevocationListsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListCACertificatesBody struct{
    Req *vsantypes.ListCACertificates `xml:"urn:vsan ListCACertificates,omitempty"`
    Res *vsantypes.ListCACertificatesResponse `xml:"urn:vsan ListCACertificatesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListCACertificatesBody) Fault() *soap.Fault { return b.Fault_ }

func ListCACertificates(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListCACertificates) (*vsantypes.ListCACertificatesResponse, error) {
  var reqBody, resBody ListCACertificatesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListFilesInGuestBody struct{
    Req *vsantypes.ListFilesInGuest `xml:"urn:vsan ListFilesInGuest,omitempty"`
    Res *vsantypes.ListFilesInGuestResponse `xml:"urn:vsan ListFilesInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListFilesInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ListFilesInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListFilesInGuest) (*vsantypes.ListFilesInGuestResponse, error) {
  var reqBody, resBody ListFilesInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListGuestAliasesBody struct{
    Req *vsantypes.ListGuestAliases `xml:"urn:vsan ListGuestAliases,omitempty"`
    Res *vsantypes.ListGuestAliasesResponse `xml:"urn:vsan ListGuestAliasesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListGuestAliasesBody) Fault() *soap.Fault { return b.Fault_ }

func ListGuestAliases(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListGuestAliases) (*vsantypes.ListGuestAliasesResponse, error) {
  var reqBody, resBody ListGuestAliasesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListGuestMappedAliasesBody struct{
    Req *vsantypes.ListGuestMappedAliases `xml:"urn:vsan ListGuestMappedAliases,omitempty"`
    Res *vsantypes.ListGuestMappedAliasesResponse `xml:"urn:vsan ListGuestMappedAliasesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListGuestMappedAliasesBody) Fault() *soap.Fault { return b.Fault_ }

func ListGuestMappedAliases(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListGuestMappedAliases) (*vsantypes.ListGuestMappedAliasesResponse, error) {
  var reqBody, resBody ListGuestMappedAliasesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListKeysBody struct{
    Req *vsantypes.ListKeys `xml:"urn:vsan ListKeys,omitempty"`
    Res *vsantypes.ListKeysResponse `xml:"urn:vsan ListKeysResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListKeysBody) Fault() *soap.Fault { return b.Fault_ }

func ListKeys(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListKeys) (*vsantypes.ListKeysResponse, error) {
  var reqBody, resBody ListKeysBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListKmipServersBody struct{
    Req *vsantypes.ListKmipServers `xml:"urn:vsan ListKmipServers,omitempty"`
    Res *vsantypes.ListKmipServersResponse `xml:"urn:vsan ListKmipServersResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListKmipServersBody) Fault() *soap.Fault { return b.Fault_ }

func ListKmipServers(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListKmipServers) (*vsantypes.ListKmipServersResponse, error) {
  var reqBody, resBody ListKmipServersBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListProcessesInGuestBody struct{
    Req *vsantypes.ListProcessesInGuest `xml:"urn:vsan ListProcessesInGuest,omitempty"`
    Res *vsantypes.ListProcessesInGuestResponse `xml:"urn:vsan ListProcessesInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListProcessesInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ListProcessesInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListProcessesInGuest) (*vsantypes.ListProcessesInGuestResponse, error) {
  var reqBody, resBody ListProcessesInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListRegistryKeysInGuestBody struct{
    Req *vsantypes.ListRegistryKeysInGuest `xml:"urn:vsan ListRegistryKeysInGuest,omitempty"`
    Res *vsantypes.ListRegistryKeysInGuestResponse `xml:"urn:vsan ListRegistryKeysInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListRegistryKeysInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ListRegistryKeysInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListRegistryKeysInGuest) (*vsantypes.ListRegistryKeysInGuestResponse, error) {
  var reqBody, resBody ListRegistryKeysInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListRegistryValuesInGuestBody struct{
    Req *vsantypes.ListRegistryValuesInGuest `xml:"urn:vsan ListRegistryValuesInGuest,omitempty"`
    Res *vsantypes.ListRegistryValuesInGuestResponse `xml:"urn:vsan ListRegistryValuesInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListRegistryValuesInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ListRegistryValuesInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListRegistryValuesInGuest) (*vsantypes.ListRegistryValuesInGuestResponse, error) {
  var reqBody, resBody ListRegistryValuesInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListSmartCardTrustAnchorsBody struct{
    Req *vsantypes.ListSmartCardTrustAnchors `xml:"urn:vsan ListSmartCardTrustAnchors,omitempty"`
    Res *vsantypes.ListSmartCardTrustAnchorsResponse `xml:"urn:vsan ListSmartCardTrustAnchorsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListSmartCardTrustAnchorsBody) Fault() *soap.Fault { return b.Fault_ }

func ListSmartCardTrustAnchors(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListSmartCardTrustAnchors) (*vsantypes.ListSmartCardTrustAnchorsResponse, error) {
  var reqBody, resBody ListSmartCardTrustAnchorsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListTagsAttachedToVStorageObjectBody struct{
    Req *vsantypes.ListTagsAttachedToVStorageObject `xml:"urn:vsan ListTagsAttachedToVStorageObject,omitempty"`
    Res *vsantypes.ListTagsAttachedToVStorageObjectResponse `xml:"urn:vsan ListTagsAttachedToVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListTagsAttachedToVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func ListTagsAttachedToVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListTagsAttachedToVStorageObject) (*vsantypes.ListTagsAttachedToVStorageObjectResponse, error) {
  var reqBody, resBody ListTagsAttachedToVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListVStorageObjectBody struct{
    Req *vsantypes.ListVStorageObject `xml:"urn:vsan ListVStorageObject,omitempty"`
    Res *vsantypes.ListVStorageObjectResponse `xml:"urn:vsan ListVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func ListVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListVStorageObject) (*vsantypes.ListVStorageObjectResponse, error) {
  var reqBody, resBody ListVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ListVStorageObjectsAttachedToTagBody struct{
    Req *vsantypes.ListVStorageObjectsAttachedToTag `xml:"urn:vsan ListVStorageObjectsAttachedToTag,omitempty"`
    Res *vsantypes.ListVStorageObjectsAttachedToTagResponse `xml:"urn:vsan ListVStorageObjectsAttachedToTagResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ListVStorageObjectsAttachedToTagBody) Fault() *soap.Fault { return b.Fault_ }

func ListVStorageObjectsAttachedToTag(ctx context.Context, r soap.RoundTripper, req *vsantypes.ListVStorageObjectsAttachedToTag) (*vsantypes.ListVStorageObjectsAttachedToTagResponse, error) {
  var reqBody, resBody ListVStorageObjectsAttachedToTagBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LogUserEventBody struct{
    Req *vsantypes.LogUserEvent `xml:"urn:vsan LogUserEvent,omitempty"`
    Res *vsantypes.LogUserEventResponse `xml:"urn:vsan LogUserEventResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LogUserEventBody) Fault() *soap.Fault { return b.Fault_ }

func LogUserEvent(ctx context.Context, r soap.RoundTripper, req *vsantypes.LogUserEvent) (*vsantypes.LogUserEventResponse, error) {
  var reqBody, resBody LogUserEventBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LoginBody struct{
    Req *vsantypes.Login `xml:"urn:vsan Login,omitempty"`
    Res *vsantypes.LoginResponse `xml:"urn:vsan LoginResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LoginBody) Fault() *soap.Fault { return b.Fault_ }

func Login(ctx context.Context, r soap.RoundTripper, req *vsantypes.Login) (*vsantypes.LoginResponse, error) {
  var reqBody, resBody LoginBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LoginBySSPIBody struct{
    Req *vsantypes.LoginBySSPI `xml:"urn:vsan LoginBySSPI,omitempty"`
    Res *vsantypes.LoginBySSPIResponse `xml:"urn:vsan LoginBySSPIResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LoginBySSPIBody) Fault() *soap.Fault { return b.Fault_ }

func LoginBySSPI(ctx context.Context, r soap.RoundTripper, req *vsantypes.LoginBySSPI) (*vsantypes.LoginBySSPIResponse, error) {
  var reqBody, resBody LoginBySSPIBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LoginByTokenBody struct{
    Req *vsantypes.LoginByToken `xml:"urn:vsan LoginByToken,omitempty"`
    Res *vsantypes.LoginByTokenResponse `xml:"urn:vsan LoginByTokenResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LoginByTokenBody) Fault() *soap.Fault { return b.Fault_ }

func LoginByToken(ctx context.Context, r soap.RoundTripper, req *vsantypes.LoginByToken) (*vsantypes.LoginByTokenResponse, error) {
  var reqBody, resBody LoginByTokenBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LoginExtensionByCertificateBody struct{
    Req *vsantypes.LoginExtensionByCertificate `xml:"urn:vsan LoginExtensionByCertificate,omitempty"`
    Res *vsantypes.LoginExtensionByCertificateResponse `xml:"urn:vsan LoginExtensionByCertificateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LoginExtensionByCertificateBody) Fault() *soap.Fault { return b.Fault_ }

func LoginExtensionByCertificate(ctx context.Context, r soap.RoundTripper, req *vsantypes.LoginExtensionByCertificate) (*vsantypes.LoginExtensionByCertificateResponse, error) {
  var reqBody, resBody LoginExtensionByCertificateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LoginExtensionBySubjectNameBody struct{
    Req *vsantypes.LoginExtensionBySubjectName `xml:"urn:vsan LoginExtensionBySubjectName,omitempty"`
    Res *vsantypes.LoginExtensionBySubjectNameResponse `xml:"urn:vsan LoginExtensionBySubjectNameResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LoginExtensionBySubjectNameBody) Fault() *soap.Fault { return b.Fault_ }

func LoginExtensionBySubjectName(ctx context.Context, r soap.RoundTripper, req *vsantypes.LoginExtensionBySubjectName) (*vsantypes.LoginExtensionBySubjectNameResponse, error) {
  var reqBody, resBody LoginExtensionBySubjectNameBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LogoutBody struct{
    Req *vsantypes.Logout `xml:"urn:vsan Logout,omitempty"`
    Res *vsantypes.LogoutResponse `xml:"urn:vsan LogoutResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LogoutBody) Fault() *soap.Fault { return b.Fault_ }

func Logout(ctx context.Context, r soap.RoundTripper, req *vsantypes.Logout) (*vsantypes.LogoutResponse, error) {
  var reqBody, resBody LogoutBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LookupDvPortGroupBody struct{
    Req *vsantypes.LookupDvPortGroup `xml:"urn:vsan LookupDvPortGroup,omitempty"`
    Res *vsantypes.LookupDvPortGroupResponse `xml:"urn:vsan LookupDvPortGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LookupDvPortGroupBody) Fault() *soap.Fault { return b.Fault_ }

func LookupDvPortGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.LookupDvPortGroup) (*vsantypes.LookupDvPortGroupResponse, error) {
  var reqBody, resBody LookupDvPortGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type LookupVmOverheadMemoryBody struct{
    Req *vsantypes.LookupVmOverheadMemory `xml:"urn:vsan LookupVmOverheadMemory,omitempty"`
    Res *vsantypes.LookupVmOverheadMemoryResponse `xml:"urn:vsan LookupVmOverheadMemoryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *LookupVmOverheadMemoryBody) Fault() *soap.Fault { return b.Fault_ }

func LookupVmOverheadMemory(ctx context.Context, r soap.RoundTripper, req *vsantypes.LookupVmOverheadMemory) (*vsantypes.LookupVmOverheadMemoryResponse, error) {
  var reqBody, resBody LookupVmOverheadMemoryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MakeDirectoryBody struct{
    Req *vsantypes.MakeDirectory `xml:"urn:vsan MakeDirectory,omitempty"`
    Res *vsantypes.MakeDirectoryResponse `xml:"urn:vsan MakeDirectoryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MakeDirectoryBody) Fault() *soap.Fault { return b.Fault_ }

func MakeDirectory(ctx context.Context, r soap.RoundTripper, req *vsantypes.MakeDirectory) (*vsantypes.MakeDirectoryResponse, error) {
  var reqBody, resBody MakeDirectoryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MakeDirectoryInGuestBody struct{
    Req *vsantypes.MakeDirectoryInGuest `xml:"urn:vsan MakeDirectoryInGuest,omitempty"`
    Res *vsantypes.MakeDirectoryInGuestResponse `xml:"urn:vsan MakeDirectoryInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MakeDirectoryInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func MakeDirectoryInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.MakeDirectoryInGuest) (*vsantypes.MakeDirectoryInGuestResponse, error) {
  var reqBody, resBody MakeDirectoryInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MakePrimaryVM_TaskBody struct{
    Req *vsantypes.MakePrimaryVM_Task `xml:"urn:vsan MakePrimaryVM_Task,omitempty"`
    Res *vsantypes.MakePrimaryVM_TaskResponse `xml:"urn:vsan MakePrimaryVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MakePrimaryVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MakePrimaryVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MakePrimaryVM_Task) (*vsantypes.MakePrimaryVM_TaskResponse, error) {
  var reqBody, resBody MakePrimaryVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkAsLocal_TaskBody struct{
    Req *vsantypes.MarkAsLocal_Task `xml:"urn:vsan MarkAsLocal_Task,omitempty"`
    Res *vsantypes.MarkAsLocal_TaskResponse `xml:"urn:vsan MarkAsLocal_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkAsLocal_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MarkAsLocal_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkAsLocal_Task) (*vsantypes.MarkAsLocal_TaskResponse, error) {
  var reqBody, resBody MarkAsLocal_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkAsNonLocal_TaskBody struct{
    Req *vsantypes.MarkAsNonLocal_Task `xml:"urn:vsan MarkAsNonLocal_Task,omitempty"`
    Res *vsantypes.MarkAsNonLocal_TaskResponse `xml:"urn:vsan MarkAsNonLocal_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkAsNonLocal_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MarkAsNonLocal_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkAsNonLocal_Task) (*vsantypes.MarkAsNonLocal_TaskResponse, error) {
  var reqBody, resBody MarkAsNonLocal_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkAsNonSsd_TaskBody struct{
    Req *vsantypes.MarkAsNonSsd_Task `xml:"urn:vsan MarkAsNonSsd_Task,omitempty"`
    Res *vsantypes.MarkAsNonSsd_TaskResponse `xml:"urn:vsan MarkAsNonSsd_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkAsNonSsd_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MarkAsNonSsd_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkAsNonSsd_Task) (*vsantypes.MarkAsNonSsd_TaskResponse, error) {
  var reqBody, resBody MarkAsNonSsd_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkAsSsd_TaskBody struct{
    Req *vsantypes.MarkAsSsd_Task `xml:"urn:vsan MarkAsSsd_Task,omitempty"`
    Res *vsantypes.MarkAsSsd_TaskResponse `xml:"urn:vsan MarkAsSsd_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkAsSsd_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MarkAsSsd_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkAsSsd_Task) (*vsantypes.MarkAsSsd_TaskResponse, error) {
  var reqBody, resBody MarkAsSsd_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkAsTemplateBody struct{
    Req *vsantypes.MarkAsTemplate `xml:"urn:vsan MarkAsTemplate,omitempty"`
    Res *vsantypes.MarkAsTemplateResponse `xml:"urn:vsan MarkAsTemplateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkAsTemplateBody) Fault() *soap.Fault { return b.Fault_ }

func MarkAsTemplate(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkAsTemplate) (*vsantypes.MarkAsTemplateResponse, error) {
  var reqBody, resBody MarkAsTemplateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkAsVirtualMachineBody struct{
    Req *vsantypes.MarkAsVirtualMachine `xml:"urn:vsan MarkAsVirtualMachine,omitempty"`
    Res *vsantypes.MarkAsVirtualMachineResponse `xml:"urn:vsan MarkAsVirtualMachineResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkAsVirtualMachineBody) Fault() *soap.Fault { return b.Fault_ }

func MarkAsVirtualMachine(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkAsVirtualMachine) (*vsantypes.MarkAsVirtualMachineResponse, error) {
  var reqBody, resBody MarkAsVirtualMachineBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkDefaultBody struct{
    Req *vsantypes.MarkDefault `xml:"urn:vsan MarkDefault,omitempty"`
    Res *vsantypes.MarkDefaultResponse `xml:"urn:vsan MarkDefaultResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkDefaultBody) Fault() *soap.Fault { return b.Fault_ }

func MarkDefault(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkDefault) (*vsantypes.MarkDefaultResponse, error) {
  var reqBody, resBody MarkDefaultBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MarkForRemovalBody struct{
    Req *vsantypes.MarkForRemoval `xml:"urn:vsan MarkForRemoval,omitempty"`
    Res *vsantypes.MarkForRemovalResponse `xml:"urn:vsan MarkForRemovalResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MarkForRemovalBody) Fault() *soap.Fault { return b.Fault_ }

func MarkForRemoval(ctx context.Context, r soap.RoundTripper, req *vsantypes.MarkForRemoval) (*vsantypes.MarkForRemovalResponse, error) {
  var reqBody, resBody MarkForRemovalBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MergeDvs_TaskBody struct{
    Req *vsantypes.MergeDvs_Task `xml:"urn:vsan MergeDvs_Task,omitempty"`
    Res *vsantypes.MergeDvs_TaskResponse `xml:"urn:vsan MergeDvs_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MergeDvs_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MergeDvs_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MergeDvs_Task) (*vsantypes.MergeDvs_TaskResponse, error) {
  var reqBody, resBody MergeDvs_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MergePermissionsBody struct{
    Req *vsantypes.MergePermissions `xml:"urn:vsan MergePermissions,omitempty"`
    Res *vsantypes.MergePermissionsResponse `xml:"urn:vsan MergePermissionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MergePermissionsBody) Fault() *soap.Fault { return b.Fault_ }

func MergePermissions(ctx context.Context, r soap.RoundTripper, req *vsantypes.MergePermissions) (*vsantypes.MergePermissionsResponse, error) {
  var reqBody, resBody MergePermissionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MigrateVM_TaskBody struct{
    Req *vsantypes.MigrateVM_Task `xml:"urn:vsan MigrateVM_Task,omitempty"`
    Res *vsantypes.MigrateVM_TaskResponse `xml:"urn:vsan MigrateVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MigrateVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MigrateVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MigrateVM_Task) (*vsantypes.MigrateVM_TaskResponse, error) {
  var reqBody, resBody MigrateVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ModifyListViewBody struct{
    Req *vsantypes.ModifyListView `xml:"urn:vsan ModifyListView,omitempty"`
    Res *vsantypes.ModifyListViewResponse `xml:"urn:vsan ModifyListViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ModifyListViewBody) Fault() *soap.Fault { return b.Fault_ }

func ModifyListView(ctx context.Context, r soap.RoundTripper, req *vsantypes.ModifyListView) (*vsantypes.ModifyListViewResponse, error) {
  var reqBody, resBody ModifyListViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MountToolsInstallerBody struct{
    Req *vsantypes.MountToolsInstaller `xml:"urn:vsan MountToolsInstaller,omitempty"`
    Res *vsantypes.MountToolsInstallerResponse `xml:"urn:vsan MountToolsInstallerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MountToolsInstallerBody) Fault() *soap.Fault { return b.Fault_ }

func MountToolsInstaller(ctx context.Context, r soap.RoundTripper, req *vsantypes.MountToolsInstaller) (*vsantypes.MountToolsInstallerResponse, error) {
  var reqBody, resBody MountToolsInstallerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MountVffsVolumeBody struct{
    Req *vsantypes.MountVffsVolume `xml:"urn:vsan MountVffsVolume,omitempty"`
    Res *vsantypes.MountVffsVolumeResponse `xml:"urn:vsan MountVffsVolumeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MountVffsVolumeBody) Fault() *soap.Fault { return b.Fault_ }

func MountVffsVolume(ctx context.Context, r soap.RoundTripper, req *vsantypes.MountVffsVolume) (*vsantypes.MountVffsVolumeResponse, error) {
  var reqBody, resBody MountVffsVolumeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MountVmfsVolumeBody struct{
    Req *vsantypes.MountVmfsVolume `xml:"urn:vsan MountVmfsVolume,omitempty"`
    Res *vsantypes.MountVmfsVolumeResponse `xml:"urn:vsan MountVmfsVolumeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MountVmfsVolumeBody) Fault() *soap.Fault { return b.Fault_ }

func MountVmfsVolume(ctx context.Context, r soap.RoundTripper, req *vsantypes.MountVmfsVolume) (*vsantypes.MountVmfsVolumeResponse, error) {
  var reqBody, resBody MountVmfsVolumeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MountVmfsVolumeEx_TaskBody struct{
    Req *vsantypes.MountVmfsVolumeEx_Task `xml:"urn:vsan MountVmfsVolumeEx_Task,omitempty"`
    Res *vsantypes.MountVmfsVolumeEx_TaskResponse `xml:"urn:vsan MountVmfsVolumeEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MountVmfsVolumeEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MountVmfsVolumeEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MountVmfsVolumeEx_Task) (*vsantypes.MountVmfsVolumeEx_TaskResponse, error) {
  var reqBody, resBody MountVmfsVolumeEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveDVPort_TaskBody struct{
    Req *vsantypes.MoveDVPort_Task `xml:"urn:vsan MoveDVPort_Task,omitempty"`
    Res *vsantypes.MoveDVPort_TaskResponse `xml:"urn:vsan MoveDVPort_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveDVPort_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MoveDVPort_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveDVPort_Task) (*vsantypes.MoveDVPort_TaskResponse, error) {
  var reqBody, resBody MoveDVPort_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveDatastoreFile_TaskBody struct{
    Req *vsantypes.MoveDatastoreFile_Task `xml:"urn:vsan MoveDatastoreFile_Task,omitempty"`
    Res *vsantypes.MoveDatastoreFile_TaskResponse `xml:"urn:vsan MoveDatastoreFile_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveDatastoreFile_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MoveDatastoreFile_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveDatastoreFile_Task) (*vsantypes.MoveDatastoreFile_TaskResponse, error) {
  var reqBody, resBody MoveDatastoreFile_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveDirectoryInGuestBody struct{
    Req *vsantypes.MoveDirectoryInGuest `xml:"urn:vsan MoveDirectoryInGuest,omitempty"`
    Res *vsantypes.MoveDirectoryInGuestResponse `xml:"urn:vsan MoveDirectoryInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveDirectoryInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func MoveDirectoryInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveDirectoryInGuest) (*vsantypes.MoveDirectoryInGuestResponse, error) {
  var reqBody, resBody MoveDirectoryInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveFileInGuestBody struct{
    Req *vsantypes.MoveFileInGuest `xml:"urn:vsan MoveFileInGuest,omitempty"`
    Res *vsantypes.MoveFileInGuestResponse `xml:"urn:vsan MoveFileInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveFileInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func MoveFileInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveFileInGuest) (*vsantypes.MoveFileInGuestResponse, error) {
  var reqBody, resBody MoveFileInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveHostInto_TaskBody struct{
    Req *vsantypes.MoveHostInto_Task `xml:"urn:vsan MoveHostInto_Task,omitempty"`
    Res *vsantypes.MoveHostInto_TaskResponse `xml:"urn:vsan MoveHostInto_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveHostInto_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MoveHostInto_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveHostInto_Task) (*vsantypes.MoveHostInto_TaskResponse, error) {
  var reqBody, resBody MoveHostInto_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveIntoFolder_TaskBody struct{
    Req *vsantypes.MoveIntoFolder_Task `xml:"urn:vsan MoveIntoFolder_Task,omitempty"`
    Res *vsantypes.MoveIntoFolder_TaskResponse `xml:"urn:vsan MoveIntoFolder_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveIntoFolder_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MoveIntoFolder_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveIntoFolder_Task) (*vsantypes.MoveIntoFolder_TaskResponse, error) {
  var reqBody, resBody MoveIntoFolder_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveIntoResourcePoolBody struct{
    Req *vsantypes.MoveIntoResourcePool `xml:"urn:vsan MoveIntoResourcePool,omitempty"`
    Res *vsantypes.MoveIntoResourcePoolResponse `xml:"urn:vsan MoveIntoResourcePoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveIntoResourcePoolBody) Fault() *soap.Fault { return b.Fault_ }

func MoveIntoResourcePool(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveIntoResourcePool) (*vsantypes.MoveIntoResourcePoolResponse, error) {
  var reqBody, resBody MoveIntoResourcePoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveInto_TaskBody struct{
    Req *vsantypes.MoveInto_Task `xml:"urn:vsan MoveInto_Task,omitempty"`
    Res *vsantypes.MoveInto_TaskResponse `xml:"urn:vsan MoveInto_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveInto_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MoveInto_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveInto_Task) (*vsantypes.MoveInto_TaskResponse, error) {
  var reqBody, resBody MoveInto_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type MoveVirtualDisk_TaskBody struct{
    Req *vsantypes.MoveVirtualDisk_Task `xml:"urn:vsan MoveVirtualDisk_Task,omitempty"`
    Res *vsantypes.MoveVirtualDisk_TaskResponse `xml:"urn:vsan MoveVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *MoveVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func MoveVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.MoveVirtualDisk_Task) (*vsantypes.MoveVirtualDisk_TaskResponse, error) {
  var reqBody, resBody MoveVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type OpenInventoryViewFolderBody struct{
    Req *vsantypes.OpenInventoryViewFolder `xml:"urn:vsan OpenInventoryViewFolder,omitempty"`
    Res *vsantypes.OpenInventoryViewFolderResponse `xml:"urn:vsan OpenInventoryViewFolderResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *OpenInventoryViewFolderBody) Fault() *soap.Fault { return b.Fault_ }

func OpenInventoryViewFolder(ctx context.Context, r soap.RoundTripper, req *vsantypes.OpenInventoryViewFolder) (*vsantypes.OpenInventoryViewFolderResponse, error) {
  var reqBody, resBody OpenInventoryViewFolderBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type OverwriteCustomizationSpecBody struct{
    Req *vsantypes.OverwriteCustomizationSpec `xml:"urn:vsan OverwriteCustomizationSpec,omitempty"`
    Res *vsantypes.OverwriteCustomizationSpecResponse `xml:"urn:vsan OverwriteCustomizationSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *OverwriteCustomizationSpecBody) Fault() *soap.Fault { return b.Fault_ }

func OverwriteCustomizationSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.OverwriteCustomizationSpec) (*vsantypes.OverwriteCustomizationSpecResponse, error) {
  var reqBody, resBody OverwriteCustomizationSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ParseDescriptorBody struct{
    Req *vsantypes.ParseDescriptor `xml:"urn:vsan ParseDescriptor,omitempty"`
    Res *vsantypes.ParseDescriptorResponse `xml:"urn:vsan ParseDescriptorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ParseDescriptorBody) Fault() *soap.Fault { return b.Fault_ }

func ParseDescriptor(ctx context.Context, r soap.RoundTripper, req *vsantypes.ParseDescriptor) (*vsantypes.ParseDescriptorResponse, error) {
  var reqBody, resBody ParseDescriptorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PerformDvsProductSpecOperation_TaskBody struct{
    Req *vsantypes.PerformDvsProductSpecOperation_Task `xml:"urn:vsan PerformDvsProductSpecOperation_Task,omitempty"`
    Res *vsantypes.PerformDvsProductSpecOperation_TaskResponse `xml:"urn:vsan PerformDvsProductSpecOperation_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PerformDvsProductSpecOperation_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PerformDvsProductSpecOperation_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PerformDvsProductSpecOperation_Task) (*vsantypes.PerformDvsProductSpecOperation_TaskResponse, error) {
  var reqBody, resBody PerformDvsProductSpecOperation_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PerformVsanUpgradeExBody struct{
    Req *vsantypes.PerformVsanUpgradeEx `xml:"urn:vsan PerformVsanUpgradeEx,omitempty"`
    Res *vsantypes.PerformVsanUpgradeExResponse `xml:"urn:vsan PerformVsanUpgradeExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PerformVsanUpgradeExBody) Fault() *soap.Fault { return b.Fault_ }

func PerformVsanUpgradeEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.PerformVsanUpgradeEx) (*vsantypes.PerformVsanUpgradeExResponse, error) {
  var reqBody, resBody PerformVsanUpgradeExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PerformVsanUpgradePreflightAsyncCheck_TaskBody struct{
    Req *vsantypes.PerformVsanUpgradePreflightAsyncCheck_Task `xml:"urn:vsan PerformVsanUpgradePreflightAsyncCheck_Task,omitempty"`
    Res *vsantypes.PerformVsanUpgradePreflightAsyncCheck_TaskResponse `xml:"urn:vsan PerformVsanUpgradePreflightAsyncCheck_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PerformVsanUpgradePreflightAsyncCheck_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PerformVsanUpgradePreflightAsyncCheck_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PerformVsanUpgradePreflightAsyncCheck_Task) (*vsantypes.PerformVsanUpgradePreflightAsyncCheck_TaskResponse, error) {
  var reqBody, resBody PerformVsanUpgradePreflightAsyncCheck_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PerformVsanUpgradePreflightCheckBody struct{
    Req *vsantypes.PerformVsanUpgradePreflightCheck `xml:"urn:vsan PerformVsanUpgradePreflightCheck,omitempty"`
    Res *vsantypes.PerformVsanUpgradePreflightCheckResponse `xml:"urn:vsan PerformVsanUpgradePreflightCheckResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PerformVsanUpgradePreflightCheckBody) Fault() *soap.Fault { return b.Fault_ }

func PerformVsanUpgradePreflightCheck(ctx context.Context, r soap.RoundTripper, req *vsantypes.PerformVsanUpgradePreflightCheck) (*vsantypes.PerformVsanUpgradePreflightCheckResponse, error) {
  var reqBody, resBody PerformVsanUpgradePreflightCheckBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PerformVsanUpgradePreflightCheckExBody struct{
    Req *vsantypes.PerformVsanUpgradePreflightCheckEx `xml:"urn:vsan PerformVsanUpgradePreflightCheckEx,omitempty"`
    Res *vsantypes.PerformVsanUpgradePreflightCheckExResponse `xml:"urn:vsan PerformVsanUpgradePreflightCheckExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PerformVsanUpgradePreflightCheckExBody) Fault() *soap.Fault { return b.Fault_ }

func PerformVsanUpgradePreflightCheckEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.PerformVsanUpgradePreflightCheckEx) (*vsantypes.PerformVsanUpgradePreflightCheckExResponse, error) {
  var reqBody, resBody PerformVsanUpgradePreflightCheckExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PerformVsanUpgrade_TaskBody struct{
    Req *vsantypes.PerformVsanUpgrade_Task `xml:"urn:vsan PerformVsanUpgrade_Task,omitempty"`
    Res *vsantypes.PerformVsanUpgrade_TaskResponse `xml:"urn:vsan PerformVsanUpgrade_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PerformVsanUpgrade_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PerformVsanUpgrade_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PerformVsanUpgrade_Task) (*vsantypes.PerformVsanUpgrade_TaskResponse, error) {
  var reqBody, resBody PerformVsanUpgrade_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PlaceVmBody struct{
    Req *vsantypes.PlaceVm `xml:"urn:vsan PlaceVm,omitempty"`
    Res *vsantypes.PlaceVmResponse `xml:"urn:vsan PlaceVmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PlaceVmBody) Fault() *soap.Fault { return b.Fault_ }

func PlaceVm(ctx context.Context, r soap.RoundTripper, req *vsantypes.PlaceVm) (*vsantypes.PlaceVmResponse, error) {
  var reqBody, resBody PlaceVmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PostEventBody struct{
    Req *vsantypes.PostEvent `xml:"urn:vsan PostEvent,omitempty"`
    Res *vsantypes.PostEventResponse `xml:"urn:vsan PostEventResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PostEventBody) Fault() *soap.Fault { return b.Fault_ }

func PostEvent(ctx context.Context, r soap.RoundTripper, req *vsantypes.PostEvent) (*vsantypes.PostEventResponse, error) {
  var reqBody, resBody PostEventBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PostHealthUpdatesBody struct{
    Req *vsantypes.PostHealthUpdates `xml:"urn:vsan PostHealthUpdates,omitempty"`
    Res *vsantypes.PostHealthUpdatesResponse `xml:"urn:vsan PostHealthUpdatesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PostHealthUpdatesBody) Fault() *soap.Fault { return b.Fault_ }

func PostHealthUpdates(ctx context.Context, r soap.RoundTripper, req *vsantypes.PostHealthUpdates) (*vsantypes.PostHealthUpdatesResponse, error) {
  var reqBody, resBody PostHealthUpdatesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PowerDownHostToStandBy_TaskBody struct{
    Req *vsantypes.PowerDownHostToStandBy_Task `xml:"urn:vsan PowerDownHostToStandBy_Task,omitempty"`
    Res *vsantypes.PowerDownHostToStandBy_TaskResponse `xml:"urn:vsan PowerDownHostToStandBy_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PowerDownHostToStandBy_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PowerDownHostToStandBy_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PowerDownHostToStandBy_Task) (*vsantypes.PowerDownHostToStandBy_TaskResponse, error) {
  var reqBody, resBody PowerDownHostToStandBy_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PowerOffVApp_TaskBody struct{
    Req *vsantypes.PowerOffVApp_Task `xml:"urn:vsan PowerOffVApp_Task,omitempty"`
    Res *vsantypes.PowerOffVApp_TaskResponse `xml:"urn:vsan PowerOffVApp_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PowerOffVApp_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PowerOffVApp_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PowerOffVApp_Task) (*vsantypes.PowerOffVApp_TaskResponse, error) {
  var reqBody, resBody PowerOffVApp_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PowerOffVM_TaskBody struct{
    Req *vsantypes.PowerOffVM_Task `xml:"urn:vsan PowerOffVM_Task,omitempty"`
    Res *vsantypes.PowerOffVM_TaskResponse `xml:"urn:vsan PowerOffVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PowerOffVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PowerOffVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PowerOffVM_Task) (*vsantypes.PowerOffVM_TaskResponse, error) {
  var reqBody, resBody PowerOffVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PowerOnMultiVM_TaskBody struct{
    Req *vsantypes.PowerOnMultiVM_Task `xml:"urn:vsan PowerOnMultiVM_Task,omitempty"`
    Res *vsantypes.PowerOnMultiVM_TaskResponse `xml:"urn:vsan PowerOnMultiVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PowerOnMultiVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PowerOnMultiVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PowerOnMultiVM_Task) (*vsantypes.PowerOnMultiVM_TaskResponse, error) {
  var reqBody, resBody PowerOnMultiVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PowerOnVApp_TaskBody struct{
    Req *vsantypes.PowerOnVApp_Task `xml:"urn:vsan PowerOnVApp_Task,omitempty"`
    Res *vsantypes.PowerOnVApp_TaskResponse `xml:"urn:vsan PowerOnVApp_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PowerOnVApp_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PowerOnVApp_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PowerOnVApp_Task) (*vsantypes.PowerOnVApp_TaskResponse, error) {
  var reqBody, resBody PowerOnVApp_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PowerOnVM_TaskBody struct{
    Req *vsantypes.PowerOnVM_Task `xml:"urn:vsan PowerOnVM_Task,omitempty"`
    Res *vsantypes.PowerOnVM_TaskResponse `xml:"urn:vsan PowerOnVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PowerOnVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PowerOnVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PowerOnVM_Task) (*vsantypes.PowerOnVM_TaskResponse, error) {
  var reqBody, resBody PowerOnVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PowerUpHostFromStandBy_TaskBody struct{
    Req *vsantypes.PowerUpHostFromStandBy_Task `xml:"urn:vsan PowerUpHostFromStandBy_Task,omitempty"`
    Res *vsantypes.PowerUpHostFromStandBy_TaskResponse `xml:"urn:vsan PowerUpHostFromStandBy_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PowerUpHostFromStandBy_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PowerUpHostFromStandBy_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PowerUpHostFromStandBy_Task) (*vsantypes.PowerUpHostFromStandBy_TaskResponse, error) {
  var reqBody, resBody PowerUpHostFromStandBy_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PrepareCryptoBody struct{
    Req *vsantypes.PrepareCrypto `xml:"urn:vsan PrepareCrypto,omitempty"`
    Res *vsantypes.PrepareCryptoResponse `xml:"urn:vsan PrepareCryptoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PrepareCryptoBody) Fault() *soap.Fault { return b.Fault_ }

func PrepareCrypto(ctx context.Context, r soap.RoundTripper, req *vsantypes.PrepareCrypto) (*vsantypes.PrepareCryptoResponse, error) {
  var reqBody, resBody PrepareCryptoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PromoteDisks_TaskBody struct{
    Req *vsantypes.PromoteDisks_Task `xml:"urn:vsan PromoteDisks_Task,omitempty"`
    Res *vsantypes.PromoteDisks_TaskResponse `xml:"urn:vsan PromoteDisks_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PromoteDisks_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PromoteDisks_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PromoteDisks_Task) (*vsantypes.PromoteDisks_TaskResponse, error) {
  var reqBody, resBody PromoteDisks_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PutUsbScanCodesBody struct{
    Req *vsantypes.PutUsbScanCodes `xml:"urn:vsan PutUsbScanCodes,omitempty"`
    Res *vsantypes.PutUsbScanCodesResponse `xml:"urn:vsan PutUsbScanCodesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PutUsbScanCodesBody) Fault() *soap.Fault { return b.Fault_ }

func PutUsbScanCodes(ctx context.Context, r soap.RoundTripper, req *vsantypes.PutUsbScanCodes) (*vsantypes.PutUsbScanCodesResponse, error) {
  var reqBody, resBody PutUsbScanCodesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAnswerFileStatusBody struct{
    Req *vsantypes.QueryAnswerFileStatus `xml:"urn:vsan QueryAnswerFileStatus,omitempty"`
    Res *vsantypes.QueryAnswerFileStatusResponse `xml:"urn:vsan QueryAnswerFileStatusResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAnswerFileStatusBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAnswerFileStatus(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAnswerFileStatus) (*vsantypes.QueryAnswerFileStatusResponse, error) {
  var reqBody, resBody QueryAnswerFileStatusBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAssignedLicensesBody struct{
    Req *vsantypes.QueryAssignedLicenses `xml:"urn:vsan QueryAssignedLicenses,omitempty"`
    Res *vsantypes.QueryAssignedLicensesResponse `xml:"urn:vsan QueryAssignedLicensesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAssignedLicensesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAssignedLicenses(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAssignedLicenses) (*vsantypes.QueryAssignedLicensesResponse, error) {
  var reqBody, resBody QueryAssignedLicensesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAvailableDisksForVmfsBody struct{
    Req *vsantypes.QueryAvailableDisksForVmfs `xml:"urn:vsan QueryAvailableDisksForVmfs,omitempty"`
    Res *vsantypes.QueryAvailableDisksForVmfsResponse `xml:"urn:vsan QueryAvailableDisksForVmfsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAvailableDisksForVmfsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAvailableDisksForVmfs(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAvailableDisksForVmfs) (*vsantypes.QueryAvailableDisksForVmfsResponse, error) {
  var reqBody, resBody QueryAvailableDisksForVmfsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAvailableDvsSpecBody struct{
    Req *vsantypes.QueryAvailableDvsSpec `xml:"urn:vsan QueryAvailableDvsSpec,omitempty"`
    Res *vsantypes.QueryAvailableDvsSpecResponse `xml:"urn:vsan QueryAvailableDvsSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAvailableDvsSpecBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAvailableDvsSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAvailableDvsSpec) (*vsantypes.QueryAvailableDvsSpecResponse, error) {
  var reqBody, resBody QueryAvailableDvsSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAvailablePartitionBody struct{
    Req *vsantypes.QueryAvailablePartition `xml:"urn:vsan QueryAvailablePartition,omitempty"`
    Res *vsantypes.QueryAvailablePartitionResponse `xml:"urn:vsan QueryAvailablePartitionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAvailablePartitionBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAvailablePartition(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAvailablePartition) (*vsantypes.QueryAvailablePartitionResponse, error) {
  var reqBody, resBody QueryAvailablePartitionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAvailablePerfMetricBody struct{
    Req *vsantypes.QueryAvailablePerfMetric `xml:"urn:vsan QueryAvailablePerfMetric,omitempty"`
    Res *vsantypes.QueryAvailablePerfMetricResponse `xml:"urn:vsan QueryAvailablePerfMetricResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAvailablePerfMetricBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAvailablePerfMetric(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAvailablePerfMetric) (*vsantypes.QueryAvailablePerfMetricResponse, error) {
  var reqBody, resBody QueryAvailablePerfMetricBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAvailableSsdsBody struct{
    Req *vsantypes.QueryAvailableSsds `xml:"urn:vsan QueryAvailableSsds,omitempty"`
    Res *vsantypes.QueryAvailableSsdsResponse `xml:"urn:vsan QueryAvailableSsdsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAvailableSsdsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAvailableSsds(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAvailableSsds) (*vsantypes.QueryAvailableSsdsResponse, error) {
  var reqBody, resBody QueryAvailableSsdsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryAvailableTimeZonesBody struct{
    Req *vsantypes.QueryAvailableTimeZones `xml:"urn:vsan QueryAvailableTimeZones,omitempty"`
    Res *vsantypes.QueryAvailableTimeZonesResponse `xml:"urn:vsan QueryAvailableTimeZonesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryAvailableTimeZonesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryAvailableTimeZones(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryAvailableTimeZones) (*vsantypes.QueryAvailableTimeZonesResponse, error) {
  var reqBody, resBody QueryAvailableTimeZonesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryBootDevicesBody struct{
    Req *vsantypes.QueryBootDevices `xml:"urn:vsan QueryBootDevices,omitempty"`
    Res *vsantypes.QueryBootDevicesResponse `xml:"urn:vsan QueryBootDevicesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryBootDevicesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryBootDevices(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryBootDevices) (*vsantypes.QueryBootDevicesResponse, error) {
  var reqBody, resBody QueryBootDevicesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryBoundVnicsBody struct{
    Req *vsantypes.QueryBoundVnics `xml:"urn:vsan QueryBoundVnics,omitempty"`
    Res *vsantypes.QueryBoundVnicsResponse `xml:"urn:vsan QueryBoundVnicsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryBoundVnicsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryBoundVnics(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryBoundVnics) (*vsantypes.QueryBoundVnicsResponse, error) {
  var reqBody, resBody QueryBoundVnicsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryCandidateNicsBody struct{
    Req *vsantypes.QueryCandidateNics `xml:"urn:vsan QueryCandidateNics,omitempty"`
    Res *vsantypes.QueryCandidateNicsResponse `xml:"urn:vsan QueryCandidateNicsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryCandidateNicsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryCandidateNics(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryCandidateNics) (*vsantypes.QueryCandidateNicsResponse, error) {
  var reqBody, resBody QueryCandidateNicsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryChangedDiskAreasBody struct{
    Req *vsantypes.QueryChangedDiskAreas `xml:"urn:vsan QueryChangedDiskAreas,omitempty"`
    Res *vsantypes.QueryChangedDiskAreasResponse `xml:"urn:vsan QueryChangedDiskAreasResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryChangedDiskAreasBody) Fault() *soap.Fault { return b.Fault_ }

func QueryChangedDiskAreas(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryChangedDiskAreas) (*vsantypes.QueryChangedDiskAreasResponse, error) {
  var reqBody, resBody QueryChangedDiskAreasBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryCmmdsBody struct{
    Req *vsantypes.QueryCmmds `xml:"urn:vsan QueryCmmds,omitempty"`
    Res *vsantypes.QueryCmmdsResponse `xml:"urn:vsan QueryCmmdsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryCmmdsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryCmmds(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryCmmds) (*vsantypes.QueryCmmdsResponse, error) {
  var reqBody, resBody QueryCmmdsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryCompatibleHostForExistingDvsBody struct{
    Req *vsantypes.QueryCompatibleHostForExistingDvs `xml:"urn:vsan QueryCompatibleHostForExistingDvs,omitempty"`
    Res *vsantypes.QueryCompatibleHostForExistingDvsResponse `xml:"urn:vsan QueryCompatibleHostForExistingDvsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryCompatibleHostForExistingDvsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryCompatibleHostForExistingDvs(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryCompatibleHostForExistingDvs) (*vsantypes.QueryCompatibleHostForExistingDvsResponse, error) {
  var reqBody, resBody QueryCompatibleHostForExistingDvsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryCompatibleHostForNewDvsBody struct{
    Req *vsantypes.QueryCompatibleHostForNewDvs `xml:"urn:vsan QueryCompatibleHostForNewDvs,omitempty"`
    Res *vsantypes.QueryCompatibleHostForNewDvsResponse `xml:"urn:vsan QueryCompatibleHostForNewDvsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryCompatibleHostForNewDvsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryCompatibleHostForNewDvs(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryCompatibleHostForNewDvs) (*vsantypes.QueryCompatibleHostForNewDvsResponse, error) {
  var reqBody, resBody QueryCompatibleHostForNewDvsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryComplianceStatusBody struct{
    Req *vsantypes.QueryComplianceStatus `xml:"urn:vsan QueryComplianceStatus,omitempty"`
    Res *vsantypes.QueryComplianceStatusResponse `xml:"urn:vsan QueryComplianceStatusResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryComplianceStatusBody) Fault() *soap.Fault { return b.Fault_ }

func QueryComplianceStatus(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryComplianceStatus) (*vsantypes.QueryComplianceStatusResponse, error) {
  var reqBody, resBody QueryComplianceStatusBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryConfigOptionBody struct{
    Req *vsantypes.QueryConfigOption `xml:"urn:vsan QueryConfigOption,omitempty"`
    Res *vsantypes.QueryConfigOptionResponse `xml:"urn:vsan QueryConfigOptionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryConfigOptionBody) Fault() *soap.Fault { return b.Fault_ }

func QueryConfigOption(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryConfigOption) (*vsantypes.QueryConfigOptionResponse, error) {
  var reqBody, resBody QueryConfigOptionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryConfigOptionDescriptorBody struct{
    Req *vsantypes.QueryConfigOptionDescriptor `xml:"urn:vsan QueryConfigOptionDescriptor,omitempty"`
    Res *vsantypes.QueryConfigOptionDescriptorResponse `xml:"urn:vsan QueryConfigOptionDescriptorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryConfigOptionDescriptorBody) Fault() *soap.Fault { return b.Fault_ }

func QueryConfigOptionDescriptor(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryConfigOptionDescriptor) (*vsantypes.QueryConfigOptionDescriptorResponse, error) {
  var reqBody, resBody QueryConfigOptionDescriptorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryConfigOptionExBody struct{
    Req *vsantypes.QueryConfigOptionEx `xml:"urn:vsan QueryConfigOptionEx,omitempty"`
    Res *vsantypes.QueryConfigOptionExResponse `xml:"urn:vsan QueryConfigOptionExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryConfigOptionExBody) Fault() *soap.Fault { return b.Fault_ }

func QueryConfigOptionEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryConfigOptionEx) (*vsantypes.QueryConfigOptionExResponse, error) {
  var reqBody, resBody QueryConfigOptionExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryConfigTargetBody struct{
    Req *vsantypes.QueryConfigTarget `xml:"urn:vsan QueryConfigTarget,omitempty"`
    Res *vsantypes.QueryConfigTargetResponse `xml:"urn:vsan QueryConfigTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryConfigTargetBody) Fault() *soap.Fault { return b.Fault_ }

func QueryConfigTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryConfigTarget) (*vsantypes.QueryConfigTargetResponse, error) {
  var reqBody, resBody QueryConfigTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryConfiguredModuleOptionStringBody struct{
    Req *vsantypes.QueryConfiguredModuleOptionString `xml:"urn:vsan QueryConfiguredModuleOptionString,omitempty"`
    Res *vsantypes.QueryConfiguredModuleOptionStringResponse `xml:"urn:vsan QueryConfiguredModuleOptionStringResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryConfiguredModuleOptionStringBody) Fault() *soap.Fault { return b.Fault_ }

func QueryConfiguredModuleOptionString(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryConfiguredModuleOptionString) (*vsantypes.QueryConfiguredModuleOptionStringResponse, error) {
  var reqBody, resBody QueryConfiguredModuleOptionStringBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryConnectionInfoBody struct{
    Req *vsantypes.QueryConnectionInfo `xml:"urn:vsan QueryConnectionInfo,omitempty"`
    Res *vsantypes.QueryConnectionInfoResponse `xml:"urn:vsan QueryConnectionInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryConnectionInfoBody) Fault() *soap.Fault { return b.Fault_ }

func QueryConnectionInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryConnectionInfo) (*vsantypes.QueryConnectionInfoResponse, error) {
  var reqBody, resBody QueryConnectionInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryConnectionInfoViaSpecBody struct{
    Req *vsantypes.QueryConnectionInfoViaSpec `xml:"urn:vsan QueryConnectionInfoViaSpec,omitempty"`
    Res *vsantypes.QueryConnectionInfoViaSpecResponse `xml:"urn:vsan QueryConnectionInfoViaSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryConnectionInfoViaSpecBody) Fault() *soap.Fault { return b.Fault_ }

func QueryConnectionInfoViaSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryConnectionInfoViaSpec) (*vsantypes.QueryConnectionInfoViaSpecResponse, error) {
  var reqBody, resBody QueryConnectionInfoViaSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDatastorePerformanceSummaryBody struct{
    Req *vsantypes.QueryDatastorePerformanceSummary `xml:"urn:vsan QueryDatastorePerformanceSummary,omitempty"`
    Res *vsantypes.QueryDatastorePerformanceSummaryResponse `xml:"urn:vsan QueryDatastorePerformanceSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDatastorePerformanceSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDatastorePerformanceSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDatastorePerformanceSummary) (*vsantypes.QueryDatastorePerformanceSummaryResponse, error) {
  var reqBody, resBody QueryDatastorePerformanceSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDateTimeBody struct{
    Req *vsantypes.QueryDateTime `xml:"urn:vsan QueryDateTime,omitempty"`
    Res *vsantypes.QueryDateTimeResponse `xml:"urn:vsan QueryDateTimeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDateTimeBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDateTime(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDateTime) (*vsantypes.QueryDateTimeResponse, error) {
  var reqBody, resBody QueryDateTimeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDescriptionsBody struct{
    Req *vsantypes.QueryDescriptions `xml:"urn:vsan QueryDescriptions,omitempty"`
    Res *vsantypes.QueryDescriptionsResponse `xml:"urn:vsan QueryDescriptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDescriptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDescriptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDescriptions) (*vsantypes.QueryDescriptionsResponse, error) {
  var reqBody, resBody QueryDescriptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDiskMappingsBody struct{
    Req *vsantypes.QueryDiskMappings `xml:"urn:vsan QueryDiskMappings,omitempty"`
    Res *vsantypes.QueryDiskMappingsResponse `xml:"urn:vsan QueryDiskMappingsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDiskMappingsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDiskMappings(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDiskMappings) (*vsantypes.QueryDiskMappingsResponse, error) {
  var reqBody, resBody QueryDiskMappingsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDisksForVsanBody struct{
    Req *vsantypes.QueryDisksForVsan `xml:"urn:vsan QueryDisksForVsan,omitempty"`
    Res *vsantypes.QueryDisksForVsanResponse `xml:"urn:vsan QueryDisksForVsanResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDisksForVsanBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDisksForVsan(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDisksForVsan) (*vsantypes.QueryDisksForVsanResponse, error) {
  var reqBody, resBody QueryDisksForVsanBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDisksUsingFilterBody struct{
    Req *vsantypes.QueryDisksUsingFilter `xml:"urn:vsan QueryDisksUsingFilter,omitempty"`
    Res *vsantypes.QueryDisksUsingFilterResponse `xml:"urn:vsan QueryDisksUsingFilterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDisksUsingFilterBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDisksUsingFilter(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDisksUsingFilter) (*vsantypes.QueryDisksUsingFilterResponse, error) {
  var reqBody, resBody QueryDisksUsingFilterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDvsByUuidBody struct{
    Req *vsantypes.QueryDvsByUuid `xml:"urn:vsan QueryDvsByUuid,omitempty"`
    Res *vsantypes.QueryDvsByUuidResponse `xml:"urn:vsan QueryDvsByUuidResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDvsByUuidBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDvsByUuid(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDvsByUuid) (*vsantypes.QueryDvsByUuidResponse, error) {
  var reqBody, resBody QueryDvsByUuidBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDvsCheckCompatibilityBody struct{
    Req *vsantypes.QueryDvsCheckCompatibility `xml:"urn:vsan QueryDvsCheckCompatibility,omitempty"`
    Res *vsantypes.QueryDvsCheckCompatibilityResponse `xml:"urn:vsan QueryDvsCheckCompatibilityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDvsCheckCompatibilityBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDvsCheckCompatibility(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDvsCheckCompatibility) (*vsantypes.QueryDvsCheckCompatibilityResponse, error) {
  var reqBody, resBody QueryDvsCheckCompatibilityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDvsCompatibleHostSpecBody struct{
    Req *vsantypes.QueryDvsCompatibleHostSpec `xml:"urn:vsan QueryDvsCompatibleHostSpec,omitempty"`
    Res *vsantypes.QueryDvsCompatibleHostSpecResponse `xml:"urn:vsan QueryDvsCompatibleHostSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDvsCompatibleHostSpecBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDvsCompatibleHostSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDvsCompatibleHostSpec) (*vsantypes.QueryDvsCompatibleHostSpecResponse, error) {
  var reqBody, resBody QueryDvsCompatibleHostSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDvsConfigTargetBody struct{
    Req *vsantypes.QueryDvsConfigTarget `xml:"urn:vsan QueryDvsConfigTarget,omitempty"`
    Res *vsantypes.QueryDvsConfigTargetResponse `xml:"urn:vsan QueryDvsConfigTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDvsConfigTargetBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDvsConfigTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDvsConfigTarget) (*vsantypes.QueryDvsConfigTargetResponse, error) {
  var reqBody, resBody QueryDvsConfigTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDvsFeatureCapabilityBody struct{
    Req *vsantypes.QueryDvsFeatureCapability `xml:"urn:vsan QueryDvsFeatureCapability,omitempty"`
    Res *vsantypes.QueryDvsFeatureCapabilityResponse `xml:"urn:vsan QueryDvsFeatureCapabilityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDvsFeatureCapabilityBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDvsFeatureCapability(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDvsFeatureCapability) (*vsantypes.QueryDvsFeatureCapabilityResponse, error) {
  var reqBody, resBody QueryDvsFeatureCapabilityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryEventsBody struct{
    Req *vsantypes.QueryEvents `xml:"urn:vsan QueryEvents,omitempty"`
    Res *vsantypes.QueryEventsResponse `xml:"urn:vsan QueryEventsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryEventsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryEvents(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryEvents) (*vsantypes.QueryEventsResponse, error) {
  var reqBody, resBody QueryEventsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryExpressionMetadataBody struct{
    Req *vsantypes.QueryExpressionMetadata `xml:"urn:vsan QueryExpressionMetadata,omitempty"`
    Res *vsantypes.QueryExpressionMetadataResponse `xml:"urn:vsan QueryExpressionMetadataResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryExpressionMetadataBody) Fault() *soap.Fault { return b.Fault_ }

func QueryExpressionMetadata(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryExpressionMetadata) (*vsantypes.QueryExpressionMetadataResponse, error) {
  var reqBody, resBody QueryExpressionMetadataBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryExtensionIpAllocationUsageBody struct{
    Req *vsantypes.QueryExtensionIpAllocationUsage `xml:"urn:vsan QueryExtensionIpAllocationUsage,omitempty"`
    Res *vsantypes.QueryExtensionIpAllocationUsageResponse `xml:"urn:vsan QueryExtensionIpAllocationUsageResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryExtensionIpAllocationUsageBody) Fault() *soap.Fault { return b.Fault_ }

func QueryExtensionIpAllocationUsage(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryExtensionIpAllocationUsage) (*vsantypes.QueryExtensionIpAllocationUsageResponse, error) {
  var reqBody, resBody QueryExtensionIpAllocationUsageBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryFaultToleranceCompatibilityBody struct{
    Req *vsantypes.QueryFaultToleranceCompatibility `xml:"urn:vsan QueryFaultToleranceCompatibility,omitempty"`
    Res *vsantypes.QueryFaultToleranceCompatibilityResponse `xml:"urn:vsan QueryFaultToleranceCompatibilityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryFaultToleranceCompatibilityBody) Fault() *soap.Fault { return b.Fault_ }

func QueryFaultToleranceCompatibility(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryFaultToleranceCompatibility) (*vsantypes.QueryFaultToleranceCompatibilityResponse, error) {
  var reqBody, resBody QueryFaultToleranceCompatibilityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryFaultToleranceCompatibilityExBody struct{
    Req *vsantypes.QueryFaultToleranceCompatibilityEx `xml:"urn:vsan QueryFaultToleranceCompatibilityEx,omitempty"`
    Res *vsantypes.QueryFaultToleranceCompatibilityExResponse `xml:"urn:vsan QueryFaultToleranceCompatibilityExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryFaultToleranceCompatibilityExBody) Fault() *soap.Fault { return b.Fault_ }

func QueryFaultToleranceCompatibilityEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryFaultToleranceCompatibilityEx) (*vsantypes.QueryFaultToleranceCompatibilityExResponse, error) {
  var reqBody, resBody QueryFaultToleranceCompatibilityExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryFilterEntitiesBody struct{
    Req *vsantypes.QueryFilterEntities `xml:"urn:vsan QueryFilterEntities,omitempty"`
    Res *vsantypes.QueryFilterEntitiesResponse `xml:"urn:vsan QueryFilterEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryFilterEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryFilterEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryFilterEntities) (*vsantypes.QueryFilterEntitiesResponse, error) {
  var reqBody, resBody QueryFilterEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryFilterInfoIdsBody struct{
    Req *vsantypes.QueryFilterInfoIds `xml:"urn:vsan QueryFilterInfoIds,omitempty"`
    Res *vsantypes.QueryFilterInfoIdsResponse `xml:"urn:vsan QueryFilterInfoIdsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryFilterInfoIdsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryFilterInfoIds(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryFilterInfoIds) (*vsantypes.QueryFilterInfoIdsResponse, error) {
  var reqBody, resBody QueryFilterInfoIdsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryFilterListBody struct{
    Req *vsantypes.QueryFilterList `xml:"urn:vsan QueryFilterList,omitempty"`
    Res *vsantypes.QueryFilterListResponse `xml:"urn:vsan QueryFilterListResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryFilterListBody) Fault() *soap.Fault { return b.Fault_ }

func QueryFilterList(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryFilterList) (*vsantypes.QueryFilterListResponse, error) {
  var reqBody, resBody QueryFilterListBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryFilterNameBody struct{
    Req *vsantypes.QueryFilterName `xml:"urn:vsan QueryFilterName,omitempty"`
    Res *vsantypes.QueryFilterNameResponse `xml:"urn:vsan QueryFilterNameResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryFilterNameBody) Fault() *soap.Fault { return b.Fault_ }

func QueryFilterName(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryFilterName) (*vsantypes.QueryFilterNameResponse, error) {
  var reqBody, resBody QueryFilterNameBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryFirmwareConfigUploadURLBody struct{
    Req *vsantypes.QueryFirmwareConfigUploadURL `xml:"urn:vsan QueryFirmwareConfigUploadURL,omitempty"`
    Res *vsantypes.QueryFirmwareConfigUploadURLResponse `xml:"urn:vsan QueryFirmwareConfigUploadURLResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryFirmwareConfigUploadURLBody) Fault() *soap.Fault { return b.Fault_ }

func QueryFirmwareConfigUploadURL(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryFirmwareConfigUploadURL) (*vsantypes.QueryFirmwareConfigUploadURLResponse, error) {
  var reqBody, resBody QueryFirmwareConfigUploadURLBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryHealthUpdateInfosBody struct{
    Req *vsantypes.QueryHealthUpdateInfos `xml:"urn:vsan QueryHealthUpdateInfos,omitempty"`
    Res *vsantypes.QueryHealthUpdateInfosResponse `xml:"urn:vsan QueryHealthUpdateInfosResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryHealthUpdateInfosBody) Fault() *soap.Fault { return b.Fault_ }

func QueryHealthUpdateInfos(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryHealthUpdateInfos) (*vsantypes.QueryHealthUpdateInfosResponse, error) {
  var reqBody, resBody QueryHealthUpdateInfosBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryHealthUpdatesBody struct{
    Req *vsantypes.QueryHealthUpdates `xml:"urn:vsan QueryHealthUpdates,omitempty"`
    Res *vsantypes.QueryHealthUpdatesResponse `xml:"urn:vsan QueryHealthUpdatesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryHealthUpdatesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryHealthUpdates(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryHealthUpdates) (*vsantypes.QueryHealthUpdatesResponse, error) {
  var reqBody, resBody QueryHealthUpdatesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryHostConnectionInfoBody struct{
    Req *vsantypes.QueryHostConnectionInfo `xml:"urn:vsan QueryHostConnectionInfo,omitempty"`
    Res *vsantypes.QueryHostConnectionInfoResponse `xml:"urn:vsan QueryHostConnectionInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryHostConnectionInfoBody) Fault() *soap.Fault { return b.Fault_ }

func QueryHostConnectionInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryHostConnectionInfo) (*vsantypes.QueryHostConnectionInfoResponse, error) {
  var reqBody, resBody QueryHostConnectionInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryHostPatch_TaskBody struct{
    Req *vsantypes.QueryHostPatch_Task `xml:"urn:vsan QueryHostPatch_Task,omitempty"`
    Res *vsantypes.QueryHostPatch_TaskResponse `xml:"urn:vsan QueryHostPatch_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryHostPatch_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func QueryHostPatch_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryHostPatch_Task) (*vsantypes.QueryHostPatch_TaskResponse, error) {
  var reqBody, resBody QueryHostPatch_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryHostProfileMetadataBody struct{
    Req *vsantypes.QueryHostProfileMetadata `xml:"urn:vsan QueryHostProfileMetadata,omitempty"`
    Res *vsantypes.QueryHostProfileMetadataResponse `xml:"urn:vsan QueryHostProfileMetadataResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryHostProfileMetadataBody) Fault() *soap.Fault { return b.Fault_ }

func QueryHostProfileMetadata(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryHostProfileMetadata) (*vsantypes.QueryHostProfileMetadataResponse, error) {
  var reqBody, resBody QueryHostProfileMetadataBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryHostStatusBody struct{
    Req *vsantypes.QueryHostStatus `xml:"urn:vsan QueryHostStatus,omitempty"`
    Res *vsantypes.QueryHostStatusResponse `xml:"urn:vsan QueryHostStatusResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryHostStatusBody) Fault() *soap.Fault { return b.Fault_ }

func QueryHostStatus(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryHostStatus) (*vsantypes.QueryHostStatusResponse, error) {
  var reqBody, resBody QueryHostStatusBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryIORMConfigOptionBody struct{
    Req *vsantypes.QueryIORMConfigOption `xml:"urn:vsan QueryIORMConfigOption,omitempty"`
    Res *vsantypes.QueryIORMConfigOptionResponse `xml:"urn:vsan QueryIORMConfigOptionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryIORMConfigOptionBody) Fault() *soap.Fault { return b.Fault_ }

func QueryIORMConfigOption(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryIORMConfigOption) (*vsantypes.QueryIORMConfigOptionResponse, error) {
  var reqBody, resBody QueryIORMConfigOptionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryIPAllocationsBody struct{
    Req *vsantypes.QueryIPAllocations `xml:"urn:vsan QueryIPAllocations,omitempty"`
    Res *vsantypes.QueryIPAllocationsResponse `xml:"urn:vsan QueryIPAllocationsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryIPAllocationsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryIPAllocations(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryIPAllocations) (*vsantypes.QueryIPAllocationsResponse, error) {
  var reqBody, resBody QueryIPAllocationsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryIoFilterInfoBody struct{
    Req *vsantypes.QueryIoFilterInfo `xml:"urn:vsan QueryIoFilterInfo,omitempty"`
    Res *vsantypes.QueryIoFilterInfoResponse `xml:"urn:vsan QueryIoFilterInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryIoFilterInfoBody) Fault() *soap.Fault { return b.Fault_ }

func QueryIoFilterInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryIoFilterInfo) (*vsantypes.QueryIoFilterInfoResponse, error) {
  var reqBody, resBody QueryIoFilterInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryIoFilterIssuesBody struct{
    Req *vsantypes.QueryIoFilterIssues `xml:"urn:vsan QueryIoFilterIssues,omitempty"`
    Res *vsantypes.QueryIoFilterIssuesResponse `xml:"urn:vsan QueryIoFilterIssuesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryIoFilterIssuesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryIoFilterIssues(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryIoFilterIssues) (*vsantypes.QueryIoFilterIssuesResponse, error) {
  var reqBody, resBody QueryIoFilterIssuesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryIpPoolsBody struct{
    Req *vsantypes.QueryIpPools `xml:"urn:vsan QueryIpPools,omitempty"`
    Res *vsantypes.QueryIpPoolsResponse `xml:"urn:vsan QueryIpPoolsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryIpPoolsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryIpPools(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryIpPools) (*vsantypes.QueryIpPoolsResponse, error) {
  var reqBody, resBody QueryIpPoolsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryLicenseSourceAvailabilityBody struct{
    Req *vsantypes.QueryLicenseSourceAvailability `xml:"urn:vsan QueryLicenseSourceAvailability,omitempty"`
    Res *vsantypes.QueryLicenseSourceAvailabilityResponse `xml:"urn:vsan QueryLicenseSourceAvailabilityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryLicenseSourceAvailabilityBody) Fault() *soap.Fault { return b.Fault_ }

func QueryLicenseSourceAvailability(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryLicenseSourceAvailability) (*vsantypes.QueryLicenseSourceAvailabilityResponse, error) {
  var reqBody, resBody QueryLicenseSourceAvailabilityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryLicenseUsageBody struct{
    Req *vsantypes.QueryLicenseUsage `xml:"urn:vsan QueryLicenseUsage,omitempty"`
    Res *vsantypes.QueryLicenseUsageResponse `xml:"urn:vsan QueryLicenseUsageResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryLicenseUsageBody) Fault() *soap.Fault { return b.Fault_ }

func QueryLicenseUsage(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryLicenseUsage) (*vsantypes.QueryLicenseUsageResponse, error) {
  var reqBody, resBody QueryLicenseUsageBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryLockdownExceptionsBody struct{
    Req *vsantypes.QueryLockdownExceptions `xml:"urn:vsan QueryLockdownExceptions,omitempty"`
    Res *vsantypes.QueryLockdownExceptionsResponse `xml:"urn:vsan QueryLockdownExceptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryLockdownExceptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryLockdownExceptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryLockdownExceptions) (*vsantypes.QueryLockdownExceptionsResponse, error) {
  var reqBody, resBody QueryLockdownExceptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryManagedByBody struct{
    Req *vsantypes.QueryManagedBy `xml:"urn:vsan QueryManagedBy,omitempty"`
    Res *vsantypes.QueryManagedByResponse `xml:"urn:vsan QueryManagedByResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryManagedByBody) Fault() *soap.Fault { return b.Fault_ }

func QueryManagedBy(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryManagedBy) (*vsantypes.QueryManagedByResponse, error) {
  var reqBody, resBody QueryManagedByBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryMemoryOverheadBody struct{
    Req *vsantypes.QueryMemoryOverhead `xml:"urn:vsan QueryMemoryOverhead,omitempty"`
    Res *vsantypes.QueryMemoryOverheadResponse `xml:"urn:vsan QueryMemoryOverheadResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryMemoryOverheadBody) Fault() *soap.Fault { return b.Fault_ }

func QueryMemoryOverhead(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryMemoryOverhead) (*vsantypes.QueryMemoryOverheadResponse, error) {
  var reqBody, resBody QueryMemoryOverheadBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryMemoryOverheadExBody struct{
    Req *vsantypes.QueryMemoryOverheadEx `xml:"urn:vsan QueryMemoryOverheadEx,omitempty"`
    Res *vsantypes.QueryMemoryOverheadExResponse `xml:"urn:vsan QueryMemoryOverheadExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryMemoryOverheadExBody) Fault() *soap.Fault { return b.Fault_ }

func QueryMemoryOverheadEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryMemoryOverheadEx) (*vsantypes.QueryMemoryOverheadExResponse, error) {
  var reqBody, resBody QueryMemoryOverheadExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryMigrationDependenciesBody struct{
    Req *vsantypes.QueryMigrationDependencies `xml:"urn:vsan QueryMigrationDependencies,omitempty"`
    Res *vsantypes.QueryMigrationDependenciesResponse `xml:"urn:vsan QueryMigrationDependenciesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryMigrationDependenciesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryMigrationDependencies(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryMigrationDependencies) (*vsantypes.QueryMigrationDependenciesResponse, error) {
  var reqBody, resBody QueryMigrationDependenciesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryModulesBody struct{
    Req *vsantypes.QueryModules `xml:"urn:vsan QueryModules,omitempty"`
    Res *vsantypes.QueryModulesResponse `xml:"urn:vsan QueryModulesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryModulesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryModules(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryModules) (*vsantypes.QueryModulesResponse, error) {
  var reqBody, resBody QueryModulesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryMonitoredEntitiesBody struct{
    Req *vsantypes.QueryMonitoredEntities `xml:"urn:vsan QueryMonitoredEntities,omitempty"`
    Res *vsantypes.QueryMonitoredEntitiesResponse `xml:"urn:vsan QueryMonitoredEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryMonitoredEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryMonitoredEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryMonitoredEntities) (*vsantypes.QueryMonitoredEntitiesResponse, error) {
  var reqBody, resBody QueryMonitoredEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryNFSUserBody struct{
    Req *vsantypes.QueryNFSUser `xml:"urn:vsan QueryNFSUser,omitempty"`
    Res *vsantypes.QueryNFSUserResponse `xml:"urn:vsan QueryNFSUserResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryNFSUserBody) Fault() *soap.Fault { return b.Fault_ }

func QueryNFSUser(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryNFSUser) (*vsantypes.QueryNFSUserResponse, error) {
  var reqBody, resBody QueryNFSUserBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryNetConfigBody struct{
    Req *vsantypes.QueryNetConfig `xml:"urn:vsan QueryNetConfig,omitempty"`
    Res *vsantypes.QueryNetConfigResponse `xml:"urn:vsan QueryNetConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryNetConfigBody) Fault() *soap.Fault { return b.Fault_ }

func QueryNetConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryNetConfig) (*vsantypes.QueryNetConfigResponse, error) {
  var reqBody, resBody QueryNetConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryNetworkHintBody struct{
    Req *vsantypes.QueryNetworkHint `xml:"urn:vsan QueryNetworkHint,omitempty"`
    Res *vsantypes.QueryNetworkHintResponse `xml:"urn:vsan QueryNetworkHintResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryNetworkHintBody) Fault() *soap.Fault { return b.Fault_ }

func QueryNetworkHint(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryNetworkHint) (*vsantypes.QueryNetworkHintResponse, error) {
  var reqBody, resBody QueryNetworkHintBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryObjectsOnPhysicalVsanDiskBody struct{
    Req *vsantypes.QueryObjectsOnPhysicalVsanDisk `xml:"urn:vsan QueryObjectsOnPhysicalVsanDisk,omitempty"`
    Res *vsantypes.QueryObjectsOnPhysicalVsanDiskResponse `xml:"urn:vsan QueryObjectsOnPhysicalVsanDiskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryObjectsOnPhysicalVsanDiskBody) Fault() *soap.Fault { return b.Fault_ }

func QueryObjectsOnPhysicalVsanDisk(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryObjectsOnPhysicalVsanDisk) (*vsantypes.QueryObjectsOnPhysicalVsanDiskResponse, error) {
  var reqBody, resBody QueryObjectsOnPhysicalVsanDiskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryOptionsBody struct{
    Req *vsantypes.QueryOptions `xml:"urn:vsan QueryOptions,omitempty"`
    Res *vsantypes.QueryOptionsResponse `xml:"urn:vsan QueryOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryOptions) (*vsantypes.QueryOptionsResponse, error) {
  var reqBody, resBody QueryOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPartitionCreateDescBody struct{
    Req *vsantypes.QueryPartitionCreateDesc `xml:"urn:vsan QueryPartitionCreateDesc,omitempty"`
    Res *vsantypes.QueryPartitionCreateDescResponse `xml:"urn:vsan QueryPartitionCreateDescResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPartitionCreateDescBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPartitionCreateDesc(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPartitionCreateDesc) (*vsantypes.QueryPartitionCreateDescResponse, error) {
  var reqBody, resBody QueryPartitionCreateDescBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPartitionCreateOptionsBody struct{
    Req *vsantypes.QueryPartitionCreateOptions `xml:"urn:vsan QueryPartitionCreateOptions,omitempty"`
    Res *vsantypes.QueryPartitionCreateOptionsResponse `xml:"urn:vsan QueryPartitionCreateOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPartitionCreateOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPartitionCreateOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPartitionCreateOptions) (*vsantypes.QueryPartitionCreateOptionsResponse, error) {
  var reqBody, resBody QueryPartitionCreateOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPathSelectionPolicyOptionsBody struct{
    Req *vsantypes.QueryPathSelectionPolicyOptions `xml:"urn:vsan QueryPathSelectionPolicyOptions,omitempty"`
    Res *vsantypes.QueryPathSelectionPolicyOptionsResponse `xml:"urn:vsan QueryPathSelectionPolicyOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPathSelectionPolicyOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPathSelectionPolicyOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPathSelectionPolicyOptions) (*vsantypes.QueryPathSelectionPolicyOptionsResponse, error) {
  var reqBody, resBody QueryPathSelectionPolicyOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPerfBody struct{
    Req *vsantypes.QueryPerf `xml:"urn:vsan QueryPerf,omitempty"`
    Res *vsantypes.QueryPerfResponse `xml:"urn:vsan QueryPerfResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPerfBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPerf(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPerf) (*vsantypes.QueryPerfResponse, error) {
  var reqBody, resBody QueryPerfBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPerfCompositeBody struct{
    Req *vsantypes.QueryPerfComposite `xml:"urn:vsan QueryPerfComposite,omitempty"`
    Res *vsantypes.QueryPerfCompositeResponse `xml:"urn:vsan QueryPerfCompositeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPerfCompositeBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPerfComposite(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPerfComposite) (*vsantypes.QueryPerfCompositeResponse, error) {
  var reqBody, resBody QueryPerfCompositeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPerfCounterBody struct{
    Req *vsantypes.QueryPerfCounter `xml:"urn:vsan QueryPerfCounter,omitempty"`
    Res *vsantypes.QueryPerfCounterResponse `xml:"urn:vsan QueryPerfCounterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPerfCounterBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPerfCounter(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPerfCounter) (*vsantypes.QueryPerfCounterResponse, error) {
  var reqBody, resBody QueryPerfCounterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPerfCounterByLevelBody struct{
    Req *vsantypes.QueryPerfCounterByLevel `xml:"urn:vsan QueryPerfCounterByLevel,omitempty"`
    Res *vsantypes.QueryPerfCounterByLevelResponse `xml:"urn:vsan QueryPerfCounterByLevelResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPerfCounterByLevelBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPerfCounterByLevel(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPerfCounterByLevel) (*vsantypes.QueryPerfCounterByLevelResponse, error) {
  var reqBody, resBody QueryPerfCounterByLevelBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPerfProviderSummaryBody struct{
    Req *vsantypes.QueryPerfProviderSummary `xml:"urn:vsan QueryPerfProviderSummary,omitempty"`
    Res *vsantypes.QueryPerfProviderSummaryResponse `xml:"urn:vsan QueryPerfProviderSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPerfProviderSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPerfProviderSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPerfProviderSummary) (*vsantypes.QueryPerfProviderSummaryResponse, error) {
  var reqBody, resBody QueryPerfProviderSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPhysicalVsanDisksBody struct{
    Req *vsantypes.QueryPhysicalVsanDisks `xml:"urn:vsan QueryPhysicalVsanDisks,omitempty"`
    Res *vsantypes.QueryPhysicalVsanDisksResponse `xml:"urn:vsan QueryPhysicalVsanDisksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPhysicalVsanDisksBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPhysicalVsanDisks(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPhysicalVsanDisks) (*vsantypes.QueryPhysicalVsanDisksResponse, error) {
  var reqBody, resBody QueryPhysicalVsanDisksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPnicStatusBody struct{
    Req *vsantypes.QueryPnicStatus `xml:"urn:vsan QueryPnicStatus,omitempty"`
    Res *vsantypes.QueryPnicStatusResponse `xml:"urn:vsan QueryPnicStatusResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPnicStatusBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPnicStatus(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPnicStatus) (*vsantypes.QueryPnicStatusResponse, error) {
  var reqBody, resBody QueryPnicStatusBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryPolicyMetadataBody struct{
    Req *vsantypes.QueryPolicyMetadata `xml:"urn:vsan QueryPolicyMetadata,omitempty"`
    Res *vsantypes.QueryPolicyMetadataResponse `xml:"urn:vsan QueryPolicyMetadataResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryPolicyMetadataBody) Fault() *soap.Fault { return b.Fault_ }

func QueryPolicyMetadata(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryPolicyMetadata) (*vsantypes.QueryPolicyMetadataResponse, error) {
  var reqBody, resBody QueryPolicyMetadataBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryProfileStructureBody struct{
    Req *vsantypes.QueryProfileStructure `xml:"urn:vsan QueryProfileStructure,omitempty"`
    Res *vsantypes.QueryProfileStructureResponse `xml:"urn:vsan QueryProfileStructureResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryProfileStructureBody) Fault() *soap.Fault { return b.Fault_ }

func QueryProfileStructure(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryProfileStructure) (*vsantypes.QueryProfileStructureResponse, error) {
  var reqBody, resBody QueryProfileStructureBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryProviderListBody struct{
    Req *vsantypes.QueryProviderList `xml:"urn:vsan QueryProviderList,omitempty"`
    Res *vsantypes.QueryProviderListResponse `xml:"urn:vsan QueryProviderListResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryProviderListBody) Fault() *soap.Fault { return b.Fault_ }

func QueryProviderList(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryProviderList) (*vsantypes.QueryProviderListResponse, error) {
  var reqBody, resBody QueryProviderListBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryProviderNameBody struct{
    Req *vsantypes.QueryProviderName `xml:"urn:vsan QueryProviderName,omitempty"`
    Res *vsantypes.QueryProviderNameResponse `xml:"urn:vsan QueryProviderNameResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryProviderNameBody) Fault() *soap.Fault { return b.Fault_ }

func QueryProviderName(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryProviderName) (*vsantypes.QueryProviderNameResponse, error) {
  var reqBody, resBody QueryProviderNameBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryResourceConfigOptionBody struct{
    Req *vsantypes.QueryResourceConfigOption `xml:"urn:vsan QueryResourceConfigOption,omitempty"`
    Res *vsantypes.QueryResourceConfigOptionResponse `xml:"urn:vsan QueryResourceConfigOptionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryResourceConfigOptionBody) Fault() *soap.Fault { return b.Fault_ }

func QueryResourceConfigOption(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryResourceConfigOption) (*vsantypes.QueryResourceConfigOptionResponse, error) {
  var reqBody, resBody QueryResourceConfigOptionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryServiceListBody struct{
    Req *vsantypes.QueryServiceList `xml:"urn:vsan QueryServiceList,omitempty"`
    Res *vsantypes.QueryServiceListResponse `xml:"urn:vsan QueryServiceListResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryServiceListBody) Fault() *soap.Fault { return b.Fault_ }

func QueryServiceList(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryServiceList) (*vsantypes.QueryServiceListResponse, error) {
  var reqBody, resBody QueryServiceListBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryStorageArrayTypePolicyOptionsBody struct{
    Req *vsantypes.QueryStorageArrayTypePolicyOptions `xml:"urn:vsan QueryStorageArrayTypePolicyOptions,omitempty"`
    Res *vsantypes.QueryStorageArrayTypePolicyOptionsResponse `xml:"urn:vsan QueryStorageArrayTypePolicyOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryStorageArrayTypePolicyOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryStorageArrayTypePolicyOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryStorageArrayTypePolicyOptions) (*vsantypes.QueryStorageArrayTypePolicyOptionsResponse, error) {
  var reqBody, resBody QueryStorageArrayTypePolicyOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QuerySupportedFeaturesBody struct{
    Req *vsantypes.QuerySupportedFeatures `xml:"urn:vsan QuerySupportedFeatures,omitempty"`
    Res *vsantypes.QuerySupportedFeaturesResponse `xml:"urn:vsan QuerySupportedFeaturesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QuerySupportedFeaturesBody) Fault() *soap.Fault { return b.Fault_ }

func QuerySupportedFeatures(ctx context.Context, r soap.RoundTripper, req *vsantypes.QuerySupportedFeatures) (*vsantypes.QuerySupportedFeaturesResponse, error) {
  var reqBody, resBody QuerySupportedFeaturesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QuerySyncingVsanObjectsBody struct{
    Req *vsantypes.QuerySyncingVsanObjects `xml:"urn:vsan QuerySyncingVsanObjects,omitempty"`
    Res *vsantypes.QuerySyncingVsanObjectsResponse `xml:"urn:vsan QuerySyncingVsanObjectsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QuerySyncingVsanObjectsBody) Fault() *soap.Fault { return b.Fault_ }

func QuerySyncingVsanObjects(ctx context.Context, r soap.RoundTripper, req *vsantypes.QuerySyncingVsanObjects) (*vsantypes.QuerySyncingVsanObjectsResponse, error) {
  var reqBody, resBody QuerySyncingVsanObjectsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QuerySystemUsersBody struct{
    Req *vsantypes.QuerySystemUsers `xml:"urn:vsan QuerySystemUsers,omitempty"`
    Res *vsantypes.QuerySystemUsersResponse `xml:"urn:vsan QuerySystemUsersResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QuerySystemUsersBody) Fault() *soap.Fault { return b.Fault_ }

func QuerySystemUsers(ctx context.Context, r soap.RoundTripper, req *vsantypes.QuerySystemUsers) (*vsantypes.QuerySystemUsersResponse, error) {
  var reqBody, resBody QuerySystemUsersBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryTargetCapabilitiesBody struct{
    Req *vsantypes.QueryTargetCapabilities `xml:"urn:vsan QueryTargetCapabilities,omitempty"`
    Res *vsantypes.QueryTargetCapabilitiesResponse `xml:"urn:vsan QueryTargetCapabilitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryTargetCapabilitiesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryTargetCapabilities(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryTargetCapabilities) (*vsantypes.QueryTargetCapabilitiesResponse, error) {
  var reqBody, resBody QueryTargetCapabilitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryTpmAttestationReportBody struct{
    Req *vsantypes.QueryTpmAttestationReport `xml:"urn:vsan QueryTpmAttestationReport,omitempty"`
    Res *vsantypes.QueryTpmAttestationReportResponse `xml:"urn:vsan QueryTpmAttestationReportResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryTpmAttestationReportBody) Fault() *soap.Fault { return b.Fault_ }

func QueryTpmAttestationReport(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryTpmAttestationReport) (*vsantypes.QueryTpmAttestationReportResponse, error) {
  var reqBody, resBody QueryTpmAttestationReportBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryUnmonitoredHostsBody struct{
    Req *vsantypes.QueryUnmonitoredHosts `xml:"urn:vsan QueryUnmonitoredHosts,omitempty"`
    Res *vsantypes.QueryUnmonitoredHostsResponse `xml:"urn:vsan QueryUnmonitoredHostsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryUnmonitoredHostsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryUnmonitoredHosts(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryUnmonitoredHosts) (*vsantypes.QueryUnmonitoredHostsResponse, error) {
  var reqBody, resBody QueryUnmonitoredHostsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryUnownedFilesBody struct{
    Req *vsantypes.QueryUnownedFiles `xml:"urn:vsan QueryUnownedFiles,omitempty"`
    Res *vsantypes.QueryUnownedFilesResponse `xml:"urn:vsan QueryUnownedFilesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryUnownedFilesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryUnownedFiles(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryUnownedFiles) (*vsantypes.QueryUnownedFilesResponse, error) {
  var reqBody, resBody QueryUnownedFilesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryUnresolvedVmfsVolumeBody struct{
    Req *vsantypes.QueryUnresolvedVmfsVolume `xml:"urn:vsan QueryUnresolvedVmfsVolume,omitempty"`
    Res *vsantypes.QueryUnresolvedVmfsVolumeResponse `xml:"urn:vsan QueryUnresolvedVmfsVolumeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryUnresolvedVmfsVolumeBody) Fault() *soap.Fault { return b.Fault_ }

func QueryUnresolvedVmfsVolume(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryUnresolvedVmfsVolume) (*vsantypes.QueryUnresolvedVmfsVolumeResponse, error) {
  var reqBody, resBody QueryUnresolvedVmfsVolumeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryUnresolvedVmfsVolumesBody struct{
    Req *vsantypes.QueryUnresolvedVmfsVolumes `xml:"urn:vsan QueryUnresolvedVmfsVolumes,omitempty"`
    Res *vsantypes.QueryUnresolvedVmfsVolumesResponse `xml:"urn:vsan QueryUnresolvedVmfsVolumesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryUnresolvedVmfsVolumesBody) Fault() *soap.Fault { return b.Fault_ }

func QueryUnresolvedVmfsVolumes(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryUnresolvedVmfsVolumes) (*vsantypes.QueryUnresolvedVmfsVolumesResponse, error) {
  var reqBody, resBody QueryUnresolvedVmfsVolumesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryUsedVlanIdInDvsBody struct{
    Req *vsantypes.QueryUsedVlanIdInDvs `xml:"urn:vsan QueryUsedVlanIdInDvs,omitempty"`
    Res *vsantypes.QueryUsedVlanIdInDvsResponse `xml:"urn:vsan QueryUsedVlanIdInDvsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryUsedVlanIdInDvsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryUsedVlanIdInDvs(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryUsedVlanIdInDvs) (*vsantypes.QueryUsedVlanIdInDvsResponse, error) {
  var reqBody, resBody QueryUsedVlanIdInDvsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVMotionCompatibilityBody struct{
    Req *vsantypes.QueryVMotionCompatibility `xml:"urn:vsan QueryVMotionCompatibility,omitempty"`
    Res *vsantypes.QueryVMotionCompatibilityResponse `xml:"urn:vsan QueryVMotionCompatibilityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVMotionCompatibilityBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVMotionCompatibility(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVMotionCompatibility) (*vsantypes.QueryVMotionCompatibilityResponse, error) {
  var reqBody, resBody QueryVMotionCompatibilityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVMotionCompatibilityEx_TaskBody struct{
    Req *vsantypes.QueryVMotionCompatibilityEx_Task `xml:"urn:vsan QueryVMotionCompatibilityEx_Task,omitempty"`
    Res *vsantypes.QueryVMotionCompatibilityEx_TaskResponse `xml:"urn:vsan QueryVMotionCompatibilityEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVMotionCompatibilityEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVMotionCompatibilityEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVMotionCompatibilityEx_Task) (*vsantypes.QueryVMotionCompatibilityEx_TaskResponse, error) {
  var reqBody, resBody QueryVMotionCompatibilityEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVirtualDiskFragmentationBody struct{
    Req *vsantypes.QueryVirtualDiskFragmentation `xml:"urn:vsan QueryVirtualDiskFragmentation,omitempty"`
    Res *vsantypes.QueryVirtualDiskFragmentationResponse `xml:"urn:vsan QueryVirtualDiskFragmentationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVirtualDiskFragmentationBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVirtualDiskFragmentation(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVirtualDiskFragmentation) (*vsantypes.QueryVirtualDiskFragmentationResponse, error) {
  var reqBody, resBody QueryVirtualDiskFragmentationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVirtualDiskGeometryBody struct{
    Req *vsantypes.QueryVirtualDiskGeometry `xml:"urn:vsan QueryVirtualDiskGeometry,omitempty"`
    Res *vsantypes.QueryVirtualDiskGeometryResponse `xml:"urn:vsan QueryVirtualDiskGeometryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVirtualDiskGeometryBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVirtualDiskGeometry(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVirtualDiskGeometry) (*vsantypes.QueryVirtualDiskGeometryResponse, error) {
  var reqBody, resBody QueryVirtualDiskGeometryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVirtualDiskUuidBody struct{
    Req *vsantypes.QueryVirtualDiskUuid `xml:"urn:vsan QueryVirtualDiskUuid,omitempty"`
    Res *vsantypes.QueryVirtualDiskUuidResponse `xml:"urn:vsan QueryVirtualDiskUuidResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVirtualDiskUuidBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVirtualDiskUuid(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVirtualDiskUuid) (*vsantypes.QueryVirtualDiskUuidResponse, error) {
  var reqBody, resBody QueryVirtualDiskUuidBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVmfsConfigOptionBody struct{
    Req *vsantypes.QueryVmfsConfigOption `xml:"urn:vsan QueryVmfsConfigOption,omitempty"`
    Res *vsantypes.QueryVmfsConfigOptionResponse `xml:"urn:vsan QueryVmfsConfigOptionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVmfsConfigOptionBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVmfsConfigOption(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVmfsConfigOption) (*vsantypes.QueryVmfsConfigOptionResponse, error) {
  var reqBody, resBody QueryVmfsConfigOptionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVmfsDatastoreCreateOptionsBody struct{
    Req *vsantypes.QueryVmfsDatastoreCreateOptions `xml:"urn:vsan QueryVmfsDatastoreCreateOptions,omitempty"`
    Res *vsantypes.QueryVmfsDatastoreCreateOptionsResponse `xml:"urn:vsan QueryVmfsDatastoreCreateOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVmfsDatastoreCreateOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVmfsDatastoreCreateOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVmfsDatastoreCreateOptions) (*vsantypes.QueryVmfsDatastoreCreateOptionsResponse, error) {
  var reqBody, resBody QueryVmfsDatastoreCreateOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVmfsDatastoreExpandOptionsBody struct{
    Req *vsantypes.QueryVmfsDatastoreExpandOptions `xml:"urn:vsan QueryVmfsDatastoreExpandOptions,omitempty"`
    Res *vsantypes.QueryVmfsDatastoreExpandOptionsResponse `xml:"urn:vsan QueryVmfsDatastoreExpandOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVmfsDatastoreExpandOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVmfsDatastoreExpandOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVmfsDatastoreExpandOptions) (*vsantypes.QueryVmfsDatastoreExpandOptionsResponse, error) {
  var reqBody, resBody QueryVmfsDatastoreExpandOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVmfsDatastoreExtendOptionsBody struct{
    Req *vsantypes.QueryVmfsDatastoreExtendOptions `xml:"urn:vsan QueryVmfsDatastoreExtendOptions,omitempty"`
    Res *vsantypes.QueryVmfsDatastoreExtendOptionsResponse `xml:"urn:vsan QueryVmfsDatastoreExtendOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVmfsDatastoreExtendOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVmfsDatastoreExtendOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVmfsDatastoreExtendOptions) (*vsantypes.QueryVmfsDatastoreExtendOptionsResponse, error) {
  var reqBody, resBody QueryVmfsDatastoreExtendOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVnicStatusBody struct{
    Req *vsantypes.QueryVnicStatus `xml:"urn:vsan QueryVnicStatus,omitempty"`
    Res *vsantypes.QueryVnicStatusResponse `xml:"urn:vsan QueryVnicStatusResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVnicStatusBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVnicStatus(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVnicStatus) (*vsantypes.QueryVnicStatusResponse, error) {
  var reqBody, resBody QueryVnicStatusBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVsanObjectUuidsByFilterBody struct{
    Req *vsantypes.QueryVsanObjectUuidsByFilter `xml:"urn:vsan QueryVsanObjectUuidsByFilter,omitempty"`
    Res *vsantypes.QueryVsanObjectUuidsByFilterResponse `xml:"urn:vsan QueryVsanObjectUuidsByFilterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVsanObjectUuidsByFilterBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVsanObjectUuidsByFilter(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVsanObjectUuidsByFilter) (*vsantypes.QueryVsanObjectUuidsByFilterResponse, error) {
  var reqBody, resBody QueryVsanObjectUuidsByFilterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVsanObjectsBody struct{
    Req *vsantypes.QueryVsanObjects `xml:"urn:vsan QueryVsanObjects,omitempty"`
    Res *vsantypes.QueryVsanObjectsResponse `xml:"urn:vsan QueryVsanObjectsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVsanObjectsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVsanObjects(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVsanObjects) (*vsantypes.QueryVsanObjectsResponse, error) {
  var reqBody, resBody QueryVsanObjectsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVsanStatisticsBody struct{
    Req *vsantypes.QueryVsanStatistics `xml:"urn:vsan QueryVsanStatistics,omitempty"`
    Res *vsantypes.QueryVsanStatisticsResponse `xml:"urn:vsan QueryVsanStatisticsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVsanStatisticsBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVsanStatistics(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVsanStatistics) (*vsantypes.QueryVsanStatisticsResponse, error) {
  var reqBody, resBody QueryVsanStatisticsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryVsanUpgradeStatusBody struct{
    Req *vsantypes.QueryVsanUpgradeStatus `xml:"urn:vsan QueryVsanUpgradeStatus,omitempty"`
    Res *vsantypes.QueryVsanUpgradeStatusResponse `xml:"urn:vsan QueryVsanUpgradeStatusResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryVsanUpgradeStatusBody) Fault() *soap.Fault { return b.Fault_ }

func QueryVsanUpgradeStatus(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryVsanUpgradeStatus) (*vsantypes.QueryVsanUpgradeStatusResponse, error) {
  var reqBody, resBody QueryVsanUpgradeStatusBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReadEnvironmentVariableInGuestBody struct{
    Req *vsantypes.ReadEnvironmentVariableInGuest `xml:"urn:vsan ReadEnvironmentVariableInGuest,omitempty"`
    Res *vsantypes.ReadEnvironmentVariableInGuestResponse `xml:"urn:vsan ReadEnvironmentVariableInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReadEnvironmentVariableInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ReadEnvironmentVariableInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReadEnvironmentVariableInGuest) (*vsantypes.ReadEnvironmentVariableInGuestResponse, error) {
  var reqBody, resBody ReadEnvironmentVariableInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReadNextEventsBody struct{
    Req *vsantypes.ReadNextEvents `xml:"urn:vsan ReadNextEvents,omitempty"`
    Res *vsantypes.ReadNextEventsResponse `xml:"urn:vsan ReadNextEventsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReadNextEventsBody) Fault() *soap.Fault { return b.Fault_ }

func ReadNextEvents(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReadNextEvents) (*vsantypes.ReadNextEventsResponse, error) {
  var reqBody, resBody ReadNextEventsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReadNextTasksBody struct{
    Req *vsantypes.ReadNextTasks `xml:"urn:vsan ReadNextTasks,omitempty"`
    Res *vsantypes.ReadNextTasksResponse `xml:"urn:vsan ReadNextTasksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReadNextTasksBody) Fault() *soap.Fault { return b.Fault_ }

func ReadNextTasks(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReadNextTasks) (*vsantypes.ReadNextTasksResponse, error) {
  var reqBody, resBody ReadNextTasksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReadPreviousEventsBody struct{
    Req *vsantypes.ReadPreviousEvents `xml:"urn:vsan ReadPreviousEvents,omitempty"`
    Res *vsantypes.ReadPreviousEventsResponse `xml:"urn:vsan ReadPreviousEventsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReadPreviousEventsBody) Fault() *soap.Fault { return b.Fault_ }

func ReadPreviousEvents(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReadPreviousEvents) (*vsantypes.ReadPreviousEventsResponse, error) {
  var reqBody, resBody ReadPreviousEventsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReadPreviousTasksBody struct{
    Req *vsantypes.ReadPreviousTasks `xml:"urn:vsan ReadPreviousTasks,omitempty"`
    Res *vsantypes.ReadPreviousTasksResponse `xml:"urn:vsan ReadPreviousTasksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReadPreviousTasksBody) Fault() *soap.Fault { return b.Fault_ }

func ReadPreviousTasks(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReadPreviousTasks) (*vsantypes.ReadPreviousTasksResponse, error) {
  var reqBody, resBody ReadPreviousTasksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RebootGuestBody struct{
    Req *vsantypes.RebootGuest `xml:"urn:vsan RebootGuest,omitempty"`
    Res *vsantypes.RebootGuestResponse `xml:"urn:vsan RebootGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RebootGuestBody) Fault() *soap.Fault { return b.Fault_ }

func RebootGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.RebootGuest) (*vsantypes.RebootGuestResponse, error) {
  var reqBody, resBody RebootGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RebootHost_TaskBody struct{
    Req *vsantypes.RebootHost_Task `xml:"urn:vsan RebootHost_Task,omitempty"`
    Res *vsantypes.RebootHost_TaskResponse `xml:"urn:vsan RebootHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RebootHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RebootHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RebootHost_Task) (*vsantypes.RebootHost_TaskResponse, error) {
  var reqBody, resBody RebootHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RecommendDatastoresBody struct{
    Req *vsantypes.RecommendDatastores `xml:"urn:vsan RecommendDatastores,omitempty"`
    Res *vsantypes.RecommendDatastoresResponse `xml:"urn:vsan RecommendDatastoresResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RecommendDatastoresBody) Fault() *soap.Fault { return b.Fault_ }

func RecommendDatastores(ctx context.Context, r soap.RoundTripper, req *vsantypes.RecommendDatastores) (*vsantypes.RecommendDatastoresResponse, error) {
  var reqBody, resBody RecommendDatastoresBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RecommendHostsForVmBody struct{
    Req *vsantypes.RecommendHostsForVm `xml:"urn:vsan RecommendHostsForVm,omitempty"`
    Res *vsantypes.RecommendHostsForVmResponse `xml:"urn:vsan RecommendHostsForVmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RecommendHostsForVmBody) Fault() *soap.Fault { return b.Fault_ }

func RecommendHostsForVm(ctx context.Context, r soap.RoundTripper, req *vsantypes.RecommendHostsForVm) (*vsantypes.RecommendHostsForVmResponse, error) {
  var reqBody, resBody RecommendHostsForVmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RecommissionVsanNode_TaskBody struct{
    Req *vsantypes.RecommissionVsanNode_Task `xml:"urn:vsan RecommissionVsanNode_Task,omitempty"`
    Res *vsantypes.RecommissionVsanNode_TaskResponse `xml:"urn:vsan RecommissionVsanNode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RecommissionVsanNode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RecommissionVsanNode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RecommissionVsanNode_Task) (*vsantypes.RecommissionVsanNode_TaskResponse, error) {
  var reqBody, resBody RecommissionVsanNode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconcileDatastoreInventory_TaskBody struct{
    Req *vsantypes.ReconcileDatastoreInventory_Task `xml:"urn:vsan ReconcileDatastoreInventory_Task,omitempty"`
    Res *vsantypes.ReconcileDatastoreInventory_TaskResponse `xml:"urn:vsan ReconcileDatastoreInventory_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconcileDatastoreInventory_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconcileDatastoreInventory_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconcileDatastoreInventory_Task) (*vsantypes.ReconcileDatastoreInventory_TaskResponse, error) {
  var reqBody, resBody ReconcileDatastoreInventory_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigVM_TaskBody struct{
    Req *vsantypes.ReconfigVM_Task `xml:"urn:vsan ReconfigVM_Task,omitempty"`
    Res *vsantypes.ReconfigVM_TaskResponse `xml:"urn:vsan ReconfigVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigVM_Task) (*vsantypes.ReconfigVM_TaskResponse, error) {
  var reqBody, resBody ReconfigVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigurationSatisfiableBody struct{
    Req *vsantypes.ReconfigurationSatisfiable `xml:"urn:vsan ReconfigurationSatisfiable,omitempty"`
    Res *vsantypes.ReconfigurationSatisfiableResponse `xml:"urn:vsan ReconfigurationSatisfiableResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigurationSatisfiableBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigurationSatisfiable(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigurationSatisfiable) (*vsantypes.ReconfigurationSatisfiableResponse, error) {
  var reqBody, resBody ReconfigurationSatisfiableBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureAlarmBody struct{
    Req *vsantypes.ReconfigureAlarm `xml:"urn:vsan ReconfigureAlarm,omitempty"`
    Res *vsantypes.ReconfigureAlarmResponse `xml:"urn:vsan ReconfigureAlarmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureAlarmBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureAlarm(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureAlarm) (*vsantypes.ReconfigureAlarmResponse, error) {
  var reqBody, resBody ReconfigureAlarmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureAutostartBody struct{
    Req *vsantypes.ReconfigureAutostart `xml:"urn:vsan ReconfigureAutostart,omitempty"`
    Res *vsantypes.ReconfigureAutostartResponse `xml:"urn:vsan ReconfigureAutostartResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureAutostartBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureAutostart(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureAutostart) (*vsantypes.ReconfigureAutostartResponse, error) {
  var reqBody, resBody ReconfigureAutostartBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureCluster_TaskBody struct{
    Req *vsantypes.ReconfigureCluster_Task `xml:"urn:vsan ReconfigureCluster_Task,omitempty"`
    Res *vsantypes.ReconfigureCluster_TaskResponse `xml:"urn:vsan ReconfigureCluster_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureCluster_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureCluster_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureCluster_Task) (*vsantypes.ReconfigureCluster_TaskResponse, error) {
  var reqBody, resBody ReconfigureCluster_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureComputeResource_TaskBody struct{
    Req *vsantypes.ReconfigureComputeResource_Task `xml:"urn:vsan ReconfigureComputeResource_Task,omitempty"`
    Res *vsantypes.ReconfigureComputeResource_TaskResponse `xml:"urn:vsan ReconfigureComputeResource_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureComputeResource_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureComputeResource_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureComputeResource_Task) (*vsantypes.ReconfigureComputeResource_TaskResponse, error) {
  var reqBody, resBody ReconfigureComputeResource_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureDVPort_TaskBody struct{
    Req *vsantypes.ReconfigureDVPort_Task `xml:"urn:vsan ReconfigureDVPort_Task,omitempty"`
    Res *vsantypes.ReconfigureDVPort_TaskResponse `xml:"urn:vsan ReconfigureDVPort_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureDVPort_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureDVPort_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureDVPort_Task) (*vsantypes.ReconfigureDVPort_TaskResponse, error) {
  var reqBody, resBody ReconfigureDVPort_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureDVPortgroup_TaskBody struct{
    Req *vsantypes.ReconfigureDVPortgroup_Task `xml:"urn:vsan ReconfigureDVPortgroup_Task,omitempty"`
    Res *vsantypes.ReconfigureDVPortgroup_TaskResponse `xml:"urn:vsan ReconfigureDVPortgroup_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureDVPortgroup_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureDVPortgroup_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureDVPortgroup_Task) (*vsantypes.ReconfigureDVPortgroup_TaskResponse, error) {
  var reqBody, resBody ReconfigureDVPortgroup_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureDatacenter_TaskBody struct{
    Req *vsantypes.ReconfigureDatacenter_Task `xml:"urn:vsan ReconfigureDatacenter_Task,omitempty"`
    Res *vsantypes.ReconfigureDatacenter_TaskResponse `xml:"urn:vsan ReconfigureDatacenter_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureDatacenter_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureDatacenter_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureDatacenter_Task) (*vsantypes.ReconfigureDatacenter_TaskResponse, error) {
  var reqBody, resBody ReconfigureDatacenter_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureDomObjectBody struct{
    Req *vsantypes.ReconfigureDomObject `xml:"urn:vsan ReconfigureDomObject,omitempty"`
    Res *vsantypes.ReconfigureDomObjectResponse `xml:"urn:vsan ReconfigureDomObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureDomObjectBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureDomObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureDomObject) (*vsantypes.ReconfigureDomObjectResponse, error) {
  var reqBody, resBody ReconfigureDomObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureDvs_TaskBody struct{
    Req *vsantypes.ReconfigureDvs_Task `xml:"urn:vsan ReconfigureDvs_Task,omitempty"`
    Res *vsantypes.ReconfigureDvs_TaskResponse `xml:"urn:vsan ReconfigureDvs_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureDvs_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureDvs_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureDvs_Task) (*vsantypes.ReconfigureDvs_TaskResponse, error) {
  var reqBody, resBody ReconfigureDvs_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureHostForDAS_TaskBody struct{
    Req *vsantypes.ReconfigureHostForDAS_Task `xml:"urn:vsan ReconfigureHostForDAS_Task,omitempty"`
    Res *vsantypes.ReconfigureHostForDAS_TaskResponse `xml:"urn:vsan ReconfigureHostForDAS_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureHostForDAS_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureHostForDAS_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureHostForDAS_Task) (*vsantypes.ReconfigureHostForDAS_TaskResponse, error) {
  var reqBody, resBody ReconfigureHostForDAS_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureScheduledTaskBody struct{
    Req *vsantypes.ReconfigureScheduledTask `xml:"urn:vsan ReconfigureScheduledTask,omitempty"`
    Res *vsantypes.ReconfigureScheduledTaskResponse `xml:"urn:vsan ReconfigureScheduledTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureScheduledTaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureScheduledTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureScheduledTask) (*vsantypes.ReconfigureScheduledTaskResponse, error) {
  var reqBody, resBody ReconfigureScheduledTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureServiceConsoleReservationBody struct{
    Req *vsantypes.ReconfigureServiceConsoleReservation `xml:"urn:vsan ReconfigureServiceConsoleReservation,omitempty"`
    Res *vsantypes.ReconfigureServiceConsoleReservationResponse `xml:"urn:vsan ReconfigureServiceConsoleReservationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureServiceConsoleReservationBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureServiceConsoleReservation(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureServiceConsoleReservation) (*vsantypes.ReconfigureServiceConsoleReservationResponse, error) {
  var reqBody, resBody ReconfigureServiceConsoleReservationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureSnmpAgentBody struct{
    Req *vsantypes.ReconfigureSnmpAgent `xml:"urn:vsan ReconfigureSnmpAgent,omitempty"`
    Res *vsantypes.ReconfigureSnmpAgentResponse `xml:"urn:vsan ReconfigureSnmpAgentResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureSnmpAgentBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureSnmpAgent(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureSnmpAgent) (*vsantypes.ReconfigureSnmpAgentResponse, error) {
  var reqBody, resBody ReconfigureSnmpAgentBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconfigureVirtualMachineReservationBody struct{
    Req *vsantypes.ReconfigureVirtualMachineReservation `xml:"urn:vsan ReconfigureVirtualMachineReservation,omitempty"`
    Res *vsantypes.ReconfigureVirtualMachineReservationResponse `xml:"urn:vsan ReconfigureVirtualMachineReservationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconfigureVirtualMachineReservationBody) Fault() *soap.Fault { return b.Fault_ }

func ReconfigureVirtualMachineReservation(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconfigureVirtualMachineReservation) (*vsantypes.ReconfigureVirtualMachineReservationResponse, error) {
  var reqBody, resBody ReconfigureVirtualMachineReservationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReconnectHost_TaskBody struct{
    Req *vsantypes.ReconnectHost_Task `xml:"urn:vsan ReconnectHost_Task,omitempty"`
    Res *vsantypes.ReconnectHost_TaskResponse `xml:"urn:vsan ReconnectHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReconnectHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReconnectHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReconnectHost_Task) (*vsantypes.ReconnectHost_TaskResponse, error) {
  var reqBody, resBody ReconnectHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RectifyDvsHost_TaskBody struct{
    Req *vsantypes.RectifyDvsHost_Task `xml:"urn:vsan RectifyDvsHost_Task,omitempty"`
    Res *vsantypes.RectifyDvsHost_TaskResponse `xml:"urn:vsan RectifyDvsHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RectifyDvsHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RectifyDvsHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RectifyDvsHost_Task) (*vsantypes.RectifyDvsHost_TaskResponse, error) {
  var reqBody, resBody RectifyDvsHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RectifyDvsOnHost_TaskBody struct{
    Req *vsantypes.RectifyDvsOnHost_Task `xml:"urn:vsan RectifyDvsOnHost_Task,omitempty"`
    Res *vsantypes.RectifyDvsOnHost_TaskResponse `xml:"urn:vsan RectifyDvsOnHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RectifyDvsOnHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RectifyDvsOnHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RectifyDvsOnHost_Task) (*vsantypes.RectifyDvsOnHost_TaskResponse, error) {
  var reqBody, resBody RectifyDvsOnHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshBody struct{
    Req *vsantypes.Refresh `xml:"urn:vsan Refresh,omitempty"`
    Res *vsantypes.RefreshResponse `xml:"urn:vsan RefreshResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshBody) Fault() *soap.Fault { return b.Fault_ }

func Refresh(ctx context.Context, r soap.RoundTripper, req *vsantypes.Refresh) (*vsantypes.RefreshResponse, error) {
  var reqBody, resBody RefreshBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshDVPortStateBody struct{
    Req *vsantypes.RefreshDVPortState `xml:"urn:vsan RefreshDVPortState,omitempty"`
    Res *vsantypes.RefreshDVPortStateResponse `xml:"urn:vsan RefreshDVPortStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshDVPortStateBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshDVPortState(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshDVPortState) (*vsantypes.RefreshDVPortStateResponse, error) {
  var reqBody, resBody RefreshDVPortStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshDatastoreBody struct{
    Req *vsantypes.RefreshDatastore `xml:"urn:vsan RefreshDatastore,omitempty"`
    Res *vsantypes.RefreshDatastoreResponse `xml:"urn:vsan RefreshDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshDatastore) (*vsantypes.RefreshDatastoreResponse, error) {
  var reqBody, resBody RefreshDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshDatastoreStorageInfoBody struct{
    Req *vsantypes.RefreshDatastoreStorageInfo `xml:"urn:vsan RefreshDatastoreStorageInfo,omitempty"`
    Res *vsantypes.RefreshDatastoreStorageInfoResponse `xml:"urn:vsan RefreshDatastoreStorageInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshDatastoreStorageInfoBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshDatastoreStorageInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshDatastoreStorageInfo) (*vsantypes.RefreshDatastoreStorageInfoResponse, error) {
  var reqBody, resBody RefreshDatastoreStorageInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshDateTimeSystemBody struct{
    Req *vsantypes.RefreshDateTimeSystem `xml:"urn:vsan RefreshDateTimeSystem,omitempty"`
    Res *vsantypes.RefreshDateTimeSystemResponse `xml:"urn:vsan RefreshDateTimeSystemResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshDateTimeSystemBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshDateTimeSystem(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshDateTimeSystem) (*vsantypes.RefreshDateTimeSystemResponse, error) {
  var reqBody, resBody RefreshDateTimeSystemBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshFirewallBody struct{
    Req *vsantypes.RefreshFirewall `xml:"urn:vsan RefreshFirewall,omitempty"`
    Res *vsantypes.RefreshFirewallResponse `xml:"urn:vsan RefreshFirewallResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshFirewallBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshFirewall(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshFirewall) (*vsantypes.RefreshFirewallResponse, error) {
  var reqBody, resBody RefreshFirewallBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshGraphicsManagerBody struct{
    Req *vsantypes.RefreshGraphicsManager `xml:"urn:vsan RefreshGraphicsManager,omitempty"`
    Res *vsantypes.RefreshGraphicsManagerResponse `xml:"urn:vsan RefreshGraphicsManagerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshGraphicsManagerBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshGraphicsManager(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshGraphicsManager) (*vsantypes.RefreshGraphicsManagerResponse, error) {
  var reqBody, resBody RefreshGraphicsManagerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshHealthStatusSystemBody struct{
    Req *vsantypes.RefreshHealthStatusSystem `xml:"urn:vsan RefreshHealthStatusSystem,omitempty"`
    Res *vsantypes.RefreshHealthStatusSystemResponse `xml:"urn:vsan RefreshHealthStatusSystemResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshHealthStatusSystemBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshHealthStatusSystem(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshHealthStatusSystem) (*vsantypes.RefreshHealthStatusSystemResponse, error) {
  var reqBody, resBody RefreshHealthStatusSystemBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshNetworkSystemBody struct{
    Req *vsantypes.RefreshNetworkSystem `xml:"urn:vsan RefreshNetworkSystem,omitempty"`
    Res *vsantypes.RefreshNetworkSystemResponse `xml:"urn:vsan RefreshNetworkSystemResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshNetworkSystemBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshNetworkSystem(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshNetworkSystem) (*vsantypes.RefreshNetworkSystemResponse, error) {
  var reqBody, resBody RefreshNetworkSystemBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshRecommendationBody struct{
    Req *vsantypes.RefreshRecommendation `xml:"urn:vsan RefreshRecommendation,omitempty"`
    Res *vsantypes.RefreshRecommendationResponse `xml:"urn:vsan RefreshRecommendationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshRecommendationBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshRecommendation(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshRecommendation) (*vsantypes.RefreshRecommendationResponse, error) {
  var reqBody, resBody RefreshRecommendationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshRuntimeBody struct{
    Req *vsantypes.RefreshRuntime `xml:"urn:vsan RefreshRuntime,omitempty"`
    Res *vsantypes.RefreshRuntimeResponse `xml:"urn:vsan RefreshRuntimeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshRuntimeBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshRuntime(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshRuntime) (*vsantypes.RefreshRuntimeResponse, error) {
  var reqBody, resBody RefreshRuntimeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshServicesBody struct{
    Req *vsantypes.RefreshServices `xml:"urn:vsan RefreshServices,omitempty"`
    Res *vsantypes.RefreshServicesResponse `xml:"urn:vsan RefreshServicesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshServicesBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshServices(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshServices) (*vsantypes.RefreshServicesResponse, error) {
  var reqBody, resBody RefreshServicesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshStorageDrsRecommendationBody struct{
    Req *vsantypes.RefreshStorageDrsRecommendation `xml:"urn:vsan RefreshStorageDrsRecommendation,omitempty"`
    Res *vsantypes.RefreshStorageDrsRecommendationResponse `xml:"urn:vsan RefreshStorageDrsRecommendationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshStorageDrsRecommendationBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshStorageDrsRecommendation(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshStorageDrsRecommendation) (*vsantypes.RefreshStorageDrsRecommendationResponse, error) {
  var reqBody, resBody RefreshStorageDrsRecommendationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshStorageInfoBody struct{
    Req *vsantypes.RefreshStorageInfo `xml:"urn:vsan RefreshStorageInfo,omitempty"`
    Res *vsantypes.RefreshStorageInfoResponse `xml:"urn:vsan RefreshStorageInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshStorageInfoBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshStorageInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshStorageInfo) (*vsantypes.RefreshStorageInfoResponse, error) {
  var reqBody, resBody RefreshStorageInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RefreshStorageSystemBody struct{
    Req *vsantypes.RefreshStorageSystem `xml:"urn:vsan RefreshStorageSystem,omitempty"`
    Res *vsantypes.RefreshStorageSystemResponse `xml:"urn:vsan RefreshStorageSystemResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RefreshStorageSystemBody) Fault() *soap.Fault { return b.Fault_ }

func RefreshStorageSystem(ctx context.Context, r soap.RoundTripper, req *vsantypes.RefreshStorageSystem) (*vsantypes.RefreshStorageSystemResponse, error) {
  var reqBody, resBody RefreshStorageSystemBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RegisterChildVM_TaskBody struct{
    Req *vsantypes.RegisterChildVM_Task `xml:"urn:vsan RegisterChildVM_Task,omitempty"`
    Res *vsantypes.RegisterChildVM_TaskResponse `xml:"urn:vsan RegisterChildVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RegisterChildVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RegisterChildVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RegisterChildVM_Task) (*vsantypes.RegisterChildVM_TaskResponse, error) {
  var reqBody, resBody RegisterChildVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RegisterDiskBody struct{
    Req *vsantypes.RegisterDisk `xml:"urn:vsan RegisterDisk,omitempty"`
    Res *vsantypes.RegisterDiskResponse `xml:"urn:vsan RegisterDiskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RegisterDiskBody) Fault() *soap.Fault { return b.Fault_ }

func RegisterDisk(ctx context.Context, r soap.RoundTripper, req *vsantypes.RegisterDisk) (*vsantypes.RegisterDiskResponse, error) {
  var reqBody, resBody RegisterDiskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RegisterExtensionBody struct{
    Req *vsantypes.RegisterExtension `xml:"urn:vsan RegisterExtension,omitempty"`
    Res *vsantypes.RegisterExtensionResponse `xml:"urn:vsan RegisterExtensionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RegisterExtensionBody) Fault() *soap.Fault { return b.Fault_ }

func RegisterExtension(ctx context.Context, r soap.RoundTripper, req *vsantypes.RegisterExtension) (*vsantypes.RegisterExtensionResponse, error) {
  var reqBody, resBody RegisterExtensionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RegisterHealthUpdateProviderBody struct{
    Req *vsantypes.RegisterHealthUpdateProvider `xml:"urn:vsan RegisterHealthUpdateProvider,omitempty"`
    Res *vsantypes.RegisterHealthUpdateProviderResponse `xml:"urn:vsan RegisterHealthUpdateProviderResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RegisterHealthUpdateProviderBody) Fault() *soap.Fault { return b.Fault_ }

func RegisterHealthUpdateProvider(ctx context.Context, r soap.RoundTripper, req *vsantypes.RegisterHealthUpdateProvider) (*vsantypes.RegisterHealthUpdateProviderResponse, error) {
  var reqBody, resBody RegisterHealthUpdateProviderBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RegisterKmipServerBody struct{
    Req *vsantypes.RegisterKmipServer `xml:"urn:vsan RegisterKmipServer,omitempty"`
    Res *vsantypes.RegisterKmipServerResponse `xml:"urn:vsan RegisterKmipServerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RegisterKmipServerBody) Fault() *soap.Fault { return b.Fault_ }

func RegisterKmipServer(ctx context.Context, r soap.RoundTripper, req *vsantypes.RegisterKmipServer) (*vsantypes.RegisterKmipServerResponse, error) {
  var reqBody, resBody RegisterKmipServerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RegisterVM_TaskBody struct{
    Req *vsantypes.RegisterVM_Task `xml:"urn:vsan RegisterVM_Task,omitempty"`
    Res *vsantypes.RegisterVM_TaskResponse `xml:"urn:vsan RegisterVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RegisterVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RegisterVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RegisterVM_Task) (*vsantypes.RegisterVM_TaskResponse, error) {
  var reqBody, resBody RegisterVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReleaseCredentialsInGuestBody struct{
    Req *vsantypes.ReleaseCredentialsInGuest `xml:"urn:vsan ReleaseCredentialsInGuest,omitempty"`
    Res *vsantypes.ReleaseCredentialsInGuestResponse `xml:"urn:vsan ReleaseCredentialsInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReleaseCredentialsInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ReleaseCredentialsInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReleaseCredentialsInGuest) (*vsantypes.ReleaseCredentialsInGuestResponse, error) {
  var reqBody, resBody ReleaseCredentialsInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReleaseIpAllocationBody struct{
    Req *vsantypes.ReleaseIpAllocation `xml:"urn:vsan ReleaseIpAllocation,omitempty"`
    Res *vsantypes.ReleaseIpAllocationResponse `xml:"urn:vsan ReleaseIpAllocationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReleaseIpAllocationBody) Fault() *soap.Fault { return b.Fault_ }

func ReleaseIpAllocation(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReleaseIpAllocation) (*vsantypes.ReleaseIpAllocationResponse, error) {
  var reqBody, resBody ReleaseIpAllocationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReleaseManagedSnapshotBody struct{
    Req *vsantypes.ReleaseManagedSnapshot `xml:"urn:vsan ReleaseManagedSnapshot,omitempty"`
    Res *vsantypes.ReleaseManagedSnapshotResponse `xml:"urn:vsan ReleaseManagedSnapshotResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReleaseManagedSnapshotBody) Fault() *soap.Fault { return b.Fault_ }

func ReleaseManagedSnapshot(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReleaseManagedSnapshot) (*vsantypes.ReleaseManagedSnapshotResponse, error) {
  var reqBody, resBody ReleaseManagedSnapshotBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReloadBody struct{
    Req *vsantypes.Reload `xml:"urn:vsan Reload,omitempty"`
    Res *vsantypes.ReloadResponse `xml:"urn:vsan ReloadResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReloadBody) Fault() *soap.Fault { return b.Fault_ }

func Reload(ctx context.Context, r soap.RoundTripper, req *vsantypes.Reload) (*vsantypes.ReloadResponse, error) {
  var reqBody, resBody ReloadBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RelocateVM_TaskBody struct{
    Req *vsantypes.RelocateVM_Task `xml:"urn:vsan RelocateVM_Task,omitempty"`
    Res *vsantypes.RelocateVM_TaskResponse `xml:"urn:vsan RelocateVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RelocateVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RelocateVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RelocateVM_Task) (*vsantypes.RelocateVM_TaskResponse, error) {
  var reqBody, resBody RelocateVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RelocateVStorageObject_TaskBody struct{
    Req *vsantypes.RelocateVStorageObject_Task `xml:"urn:vsan RelocateVStorageObject_Task,omitempty"`
    Res *vsantypes.RelocateVStorageObject_TaskResponse `xml:"urn:vsan RelocateVStorageObject_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RelocateVStorageObject_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RelocateVStorageObject_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RelocateVStorageObject_Task) (*vsantypes.RelocateVStorageObject_TaskResponse, error) {
  var reqBody, resBody RelocateVStorageObject_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveAlarmBody struct{
    Req *vsantypes.RemoveAlarm `xml:"urn:vsan RemoveAlarm,omitempty"`
    Res *vsantypes.RemoveAlarmResponse `xml:"urn:vsan RemoveAlarmResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveAlarmBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveAlarm(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveAlarm) (*vsantypes.RemoveAlarmResponse, error) {
  var reqBody, resBody RemoveAlarmBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveAllSnapshots_TaskBody struct{
    Req *vsantypes.RemoveAllSnapshots_Task `xml:"urn:vsan RemoveAllSnapshots_Task,omitempty"`
    Res *vsantypes.RemoveAllSnapshots_TaskResponse `xml:"urn:vsan RemoveAllSnapshots_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveAllSnapshots_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveAllSnapshots_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveAllSnapshots_Task) (*vsantypes.RemoveAllSnapshots_TaskResponse, error) {
  var reqBody, resBody RemoveAllSnapshots_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveAssignedLicenseBody struct{
    Req *vsantypes.RemoveAssignedLicense `xml:"urn:vsan RemoveAssignedLicense,omitempty"`
    Res *vsantypes.RemoveAssignedLicenseResponse `xml:"urn:vsan RemoveAssignedLicenseResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveAssignedLicenseBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveAssignedLicense(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveAssignedLicense) (*vsantypes.RemoveAssignedLicenseResponse, error) {
  var reqBody, resBody RemoveAssignedLicenseBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveAuthorizationRoleBody struct{
    Req *vsantypes.RemoveAuthorizationRole `xml:"urn:vsan RemoveAuthorizationRole,omitempty"`
    Res *vsantypes.RemoveAuthorizationRoleResponse `xml:"urn:vsan RemoveAuthorizationRoleResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveAuthorizationRoleBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveAuthorizationRole(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveAuthorizationRole) (*vsantypes.RemoveAuthorizationRoleResponse, error) {
  var reqBody, resBody RemoveAuthorizationRoleBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveCustomFieldDefBody struct{
    Req *vsantypes.RemoveCustomFieldDef `xml:"urn:vsan RemoveCustomFieldDef,omitempty"`
    Res *vsantypes.RemoveCustomFieldDefResponse `xml:"urn:vsan RemoveCustomFieldDefResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveCustomFieldDefBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveCustomFieldDef(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveCustomFieldDef) (*vsantypes.RemoveCustomFieldDefResponse, error) {
  var reqBody, resBody RemoveCustomFieldDefBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveDatastoreBody struct{
    Req *vsantypes.RemoveDatastore `xml:"urn:vsan RemoveDatastore,omitempty"`
    Res *vsantypes.RemoveDatastoreResponse `xml:"urn:vsan RemoveDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveDatastore) (*vsantypes.RemoveDatastoreResponse, error) {
  var reqBody, resBody RemoveDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveDatastoreEx_TaskBody struct{
    Req *vsantypes.RemoveDatastoreEx_Task `xml:"urn:vsan RemoveDatastoreEx_Task,omitempty"`
    Res *vsantypes.RemoveDatastoreEx_TaskResponse `xml:"urn:vsan RemoveDatastoreEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveDatastoreEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveDatastoreEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveDatastoreEx_Task) (*vsantypes.RemoveDatastoreEx_TaskResponse, error) {
  var reqBody, resBody RemoveDatastoreEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveDiskMapping_TaskBody struct{
    Req *vsantypes.RemoveDiskMapping_Task `xml:"urn:vsan RemoveDiskMapping_Task,omitempty"`
    Res *vsantypes.RemoveDiskMapping_TaskResponse `xml:"urn:vsan RemoveDiskMapping_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveDiskMapping_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveDiskMapping_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveDiskMapping_Task) (*vsantypes.RemoveDiskMapping_TaskResponse, error) {
  var reqBody, resBody RemoveDiskMapping_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveDisk_TaskBody struct{
    Req *vsantypes.RemoveDisk_Task `xml:"urn:vsan RemoveDisk_Task,omitempty"`
    Res *vsantypes.RemoveDisk_TaskResponse `xml:"urn:vsan RemoveDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveDisk_Task) (*vsantypes.RemoveDisk_TaskResponse, error) {
  var reqBody, resBody RemoveDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveEntityPermissionBody struct{
    Req *vsantypes.RemoveEntityPermission `xml:"urn:vsan RemoveEntityPermission,omitempty"`
    Res *vsantypes.RemoveEntityPermissionResponse `xml:"urn:vsan RemoveEntityPermissionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveEntityPermissionBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveEntityPermission(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveEntityPermission) (*vsantypes.RemoveEntityPermissionResponse, error) {
  var reqBody, resBody RemoveEntityPermissionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveFilterBody struct{
    Req *vsantypes.RemoveFilter `xml:"urn:vsan RemoveFilter,omitempty"`
    Res *vsantypes.RemoveFilterResponse `xml:"urn:vsan RemoveFilterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveFilterBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveFilter(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveFilter) (*vsantypes.RemoveFilterResponse, error) {
  var reqBody, resBody RemoveFilterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveFilterEntitiesBody struct{
    Req *vsantypes.RemoveFilterEntities `xml:"urn:vsan RemoveFilterEntities,omitempty"`
    Res *vsantypes.RemoveFilterEntitiesResponse `xml:"urn:vsan RemoveFilterEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveFilterEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveFilterEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveFilterEntities) (*vsantypes.RemoveFilterEntitiesResponse, error) {
  var reqBody, resBody RemoveFilterEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveGroupBody struct{
    Req *vsantypes.RemoveGroup `xml:"urn:vsan RemoveGroup,omitempty"`
    Res *vsantypes.RemoveGroupResponse `xml:"urn:vsan RemoveGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveGroupBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveGroup) (*vsantypes.RemoveGroupResponse, error) {
  var reqBody, resBody RemoveGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveGuestAliasBody struct{
    Req *vsantypes.RemoveGuestAlias `xml:"urn:vsan RemoveGuestAlias,omitempty"`
    Res *vsantypes.RemoveGuestAliasResponse `xml:"urn:vsan RemoveGuestAliasResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveGuestAliasBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveGuestAlias(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveGuestAlias) (*vsantypes.RemoveGuestAliasResponse, error) {
  var reqBody, resBody RemoveGuestAliasBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveGuestAliasByCertBody struct{
    Req *vsantypes.RemoveGuestAliasByCert `xml:"urn:vsan RemoveGuestAliasByCert,omitempty"`
    Res *vsantypes.RemoveGuestAliasByCertResponse `xml:"urn:vsan RemoveGuestAliasByCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveGuestAliasByCertBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveGuestAliasByCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveGuestAliasByCert) (*vsantypes.RemoveGuestAliasByCertResponse, error) {
  var reqBody, resBody RemoveGuestAliasByCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveInternetScsiSendTargetsBody struct{
    Req *vsantypes.RemoveInternetScsiSendTargets `xml:"urn:vsan RemoveInternetScsiSendTargets,omitempty"`
    Res *vsantypes.RemoveInternetScsiSendTargetsResponse `xml:"urn:vsan RemoveInternetScsiSendTargetsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveInternetScsiSendTargetsBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveInternetScsiSendTargets(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveInternetScsiSendTargets) (*vsantypes.RemoveInternetScsiSendTargetsResponse, error) {
  var reqBody, resBody RemoveInternetScsiSendTargetsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveInternetScsiStaticTargetsBody struct{
    Req *vsantypes.RemoveInternetScsiStaticTargets `xml:"urn:vsan RemoveInternetScsiStaticTargets,omitempty"`
    Res *vsantypes.RemoveInternetScsiStaticTargetsResponse `xml:"urn:vsan RemoveInternetScsiStaticTargetsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveInternetScsiStaticTargetsBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveInternetScsiStaticTargets(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveInternetScsiStaticTargets) (*vsantypes.RemoveInternetScsiStaticTargetsResponse, error) {
  var reqBody, resBody RemoveInternetScsiStaticTargetsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveKeyBody struct{
    Req *vsantypes.RemoveKey `xml:"urn:vsan RemoveKey,omitempty"`
    Res *vsantypes.RemoveKeyResponse `xml:"urn:vsan RemoveKeyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveKeyBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveKey(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveKey) (*vsantypes.RemoveKeyResponse, error) {
  var reqBody, resBody RemoveKeyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveKeysBody struct{
    Req *vsantypes.RemoveKeys `xml:"urn:vsan RemoveKeys,omitempty"`
    Res *vsantypes.RemoveKeysResponse `xml:"urn:vsan RemoveKeysResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveKeysBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveKeys(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveKeys) (*vsantypes.RemoveKeysResponse, error) {
  var reqBody, resBody RemoveKeysBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveKmipServerBody struct{
    Req *vsantypes.RemoveKmipServer `xml:"urn:vsan RemoveKmipServer,omitempty"`
    Res *vsantypes.RemoveKmipServerResponse `xml:"urn:vsan RemoveKmipServerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveKmipServerBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveKmipServer(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveKmipServer) (*vsantypes.RemoveKmipServerResponse, error) {
  var reqBody, resBody RemoveKmipServerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveLicenseBody struct{
    Req *vsantypes.RemoveLicense `xml:"urn:vsan RemoveLicense,omitempty"`
    Res *vsantypes.RemoveLicenseResponse `xml:"urn:vsan RemoveLicenseResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveLicenseBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveLicense(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveLicense) (*vsantypes.RemoveLicenseResponse, error) {
  var reqBody, resBody RemoveLicenseBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveLicenseLabelBody struct{
    Req *vsantypes.RemoveLicenseLabel `xml:"urn:vsan RemoveLicenseLabel,omitempty"`
    Res *vsantypes.RemoveLicenseLabelResponse `xml:"urn:vsan RemoveLicenseLabelResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveLicenseLabelBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveLicenseLabel(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveLicenseLabel) (*vsantypes.RemoveLicenseLabelResponse, error) {
  var reqBody, resBody RemoveLicenseLabelBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveMonitoredEntitiesBody struct{
    Req *vsantypes.RemoveMonitoredEntities `xml:"urn:vsan RemoveMonitoredEntities,omitempty"`
    Res *vsantypes.RemoveMonitoredEntitiesResponse `xml:"urn:vsan RemoveMonitoredEntitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveMonitoredEntitiesBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveMonitoredEntities(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveMonitoredEntities) (*vsantypes.RemoveMonitoredEntitiesResponse, error) {
  var reqBody, resBody RemoveMonitoredEntitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveNetworkResourcePoolBody struct{
    Req *vsantypes.RemoveNetworkResourcePool `xml:"urn:vsan RemoveNetworkResourcePool,omitempty"`
    Res *vsantypes.RemoveNetworkResourcePoolResponse `xml:"urn:vsan RemoveNetworkResourcePoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveNetworkResourcePoolBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveNetworkResourcePool(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveNetworkResourcePool) (*vsantypes.RemoveNetworkResourcePoolResponse, error) {
  var reqBody, resBody RemoveNetworkResourcePoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemovePerfIntervalBody struct{
    Req *vsantypes.RemovePerfInterval `xml:"urn:vsan RemovePerfInterval,omitempty"`
    Res *vsantypes.RemovePerfIntervalResponse `xml:"urn:vsan RemovePerfIntervalResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemovePerfIntervalBody) Fault() *soap.Fault { return b.Fault_ }

func RemovePerfInterval(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemovePerfInterval) (*vsantypes.RemovePerfIntervalResponse, error) {
  var reqBody, resBody RemovePerfIntervalBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemovePortGroupBody struct{
    Req *vsantypes.RemovePortGroup `xml:"urn:vsan RemovePortGroup,omitempty"`
    Res *vsantypes.RemovePortGroupResponse `xml:"urn:vsan RemovePortGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemovePortGroupBody) Fault() *soap.Fault { return b.Fault_ }

func RemovePortGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemovePortGroup) (*vsantypes.RemovePortGroupResponse, error) {
  var reqBody, resBody RemovePortGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveScheduledTaskBody struct{
    Req *vsantypes.RemoveScheduledTask `xml:"urn:vsan RemoveScheduledTask,omitempty"`
    Res *vsantypes.RemoveScheduledTaskResponse `xml:"urn:vsan RemoveScheduledTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveScheduledTaskBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveScheduledTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveScheduledTask) (*vsantypes.RemoveScheduledTaskResponse, error) {
  var reqBody, resBody RemoveScheduledTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveServiceConsoleVirtualNicBody struct{
    Req *vsantypes.RemoveServiceConsoleVirtualNic `xml:"urn:vsan RemoveServiceConsoleVirtualNic,omitempty"`
    Res *vsantypes.RemoveServiceConsoleVirtualNicResponse `xml:"urn:vsan RemoveServiceConsoleVirtualNicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveServiceConsoleVirtualNicBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveServiceConsoleVirtualNic(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveServiceConsoleVirtualNic) (*vsantypes.RemoveServiceConsoleVirtualNicResponse, error) {
  var reqBody, resBody RemoveServiceConsoleVirtualNicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveSmartCardTrustAnchorBody struct{
    Req *vsantypes.RemoveSmartCardTrustAnchor `xml:"urn:vsan RemoveSmartCardTrustAnchor,omitempty"`
    Res *vsantypes.RemoveSmartCardTrustAnchorResponse `xml:"urn:vsan RemoveSmartCardTrustAnchorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveSmartCardTrustAnchorBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveSmartCardTrustAnchor(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveSmartCardTrustAnchor) (*vsantypes.RemoveSmartCardTrustAnchorResponse, error) {
  var reqBody, resBody RemoveSmartCardTrustAnchorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveSmartCardTrustAnchorByFingerprintBody struct{
    Req *vsantypes.RemoveSmartCardTrustAnchorByFingerprint `xml:"urn:vsan RemoveSmartCardTrustAnchorByFingerprint,omitempty"`
    Res *vsantypes.RemoveSmartCardTrustAnchorByFingerprintResponse `xml:"urn:vsan RemoveSmartCardTrustAnchorByFingerprintResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveSmartCardTrustAnchorByFingerprintBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveSmartCardTrustAnchorByFingerprint(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveSmartCardTrustAnchorByFingerprint) (*vsantypes.RemoveSmartCardTrustAnchorByFingerprintResponse, error) {
  var reqBody, resBody RemoveSmartCardTrustAnchorByFingerprintBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveSnapshot_TaskBody struct{
    Req *vsantypes.RemoveSnapshot_Task `xml:"urn:vsan RemoveSnapshot_Task,omitempty"`
    Res *vsantypes.RemoveSnapshot_TaskResponse `xml:"urn:vsan RemoveSnapshot_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveSnapshot_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveSnapshot_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveSnapshot_Task) (*vsantypes.RemoveSnapshot_TaskResponse, error) {
  var reqBody, resBody RemoveSnapshot_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveUserBody struct{
    Req *vsantypes.RemoveUser `xml:"urn:vsan RemoveUser,omitempty"`
    Res *vsantypes.RemoveUserResponse `xml:"urn:vsan RemoveUserResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveUserBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveUser(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveUser) (*vsantypes.RemoveUserResponse, error) {
  var reqBody, resBody RemoveUserBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveVirtualNicBody struct{
    Req *vsantypes.RemoveVirtualNic `xml:"urn:vsan RemoveVirtualNic,omitempty"`
    Res *vsantypes.RemoveVirtualNicResponse `xml:"urn:vsan RemoveVirtualNicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveVirtualNicBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveVirtualNic(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveVirtualNic) (*vsantypes.RemoveVirtualNicResponse, error) {
  var reqBody, resBody RemoveVirtualNicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RemoveVirtualSwitchBody struct{
    Req *vsantypes.RemoveVirtualSwitch `xml:"urn:vsan RemoveVirtualSwitch,omitempty"`
    Res *vsantypes.RemoveVirtualSwitchResponse `xml:"urn:vsan RemoveVirtualSwitchResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RemoveVirtualSwitchBody) Fault() *soap.Fault { return b.Fault_ }

func RemoveVirtualSwitch(ctx context.Context, r soap.RoundTripper, req *vsantypes.RemoveVirtualSwitch) (*vsantypes.RemoveVirtualSwitchResponse, error) {
  var reqBody, resBody RemoveVirtualSwitchBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RenameCustomFieldDefBody struct{
    Req *vsantypes.RenameCustomFieldDef `xml:"urn:vsan RenameCustomFieldDef,omitempty"`
    Res *vsantypes.RenameCustomFieldDefResponse `xml:"urn:vsan RenameCustomFieldDefResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RenameCustomFieldDefBody) Fault() *soap.Fault { return b.Fault_ }

func RenameCustomFieldDef(ctx context.Context, r soap.RoundTripper, req *vsantypes.RenameCustomFieldDef) (*vsantypes.RenameCustomFieldDefResponse, error) {
  var reqBody, resBody RenameCustomFieldDefBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RenameCustomizationSpecBody struct{
    Req *vsantypes.RenameCustomizationSpec `xml:"urn:vsan RenameCustomizationSpec,omitempty"`
    Res *vsantypes.RenameCustomizationSpecResponse `xml:"urn:vsan RenameCustomizationSpecResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RenameCustomizationSpecBody) Fault() *soap.Fault { return b.Fault_ }

func RenameCustomizationSpec(ctx context.Context, r soap.RoundTripper, req *vsantypes.RenameCustomizationSpec) (*vsantypes.RenameCustomizationSpecResponse, error) {
  var reqBody, resBody RenameCustomizationSpecBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RenameDatastoreBody struct{
    Req *vsantypes.RenameDatastore `xml:"urn:vsan RenameDatastore,omitempty"`
    Res *vsantypes.RenameDatastoreResponse `xml:"urn:vsan RenameDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RenameDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func RenameDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.RenameDatastore) (*vsantypes.RenameDatastoreResponse, error) {
  var reqBody, resBody RenameDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RenameSnapshotBody struct{
    Req *vsantypes.RenameSnapshot `xml:"urn:vsan RenameSnapshot,omitempty"`
    Res *vsantypes.RenameSnapshotResponse `xml:"urn:vsan RenameSnapshotResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RenameSnapshotBody) Fault() *soap.Fault { return b.Fault_ }

func RenameSnapshot(ctx context.Context, r soap.RoundTripper, req *vsantypes.RenameSnapshot) (*vsantypes.RenameSnapshotResponse, error) {
  var reqBody, resBody RenameSnapshotBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RenameVStorageObjectBody struct{
    Req *vsantypes.RenameVStorageObject `xml:"urn:vsan RenameVStorageObject,omitempty"`
    Res *vsantypes.RenameVStorageObjectResponse `xml:"urn:vsan RenameVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RenameVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func RenameVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.RenameVStorageObject) (*vsantypes.RenameVStorageObjectResponse, error) {
  var reqBody, resBody RenameVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type Rename_TaskBody struct{
    Req *vsantypes.Rename_Task `xml:"urn:vsan Rename_Task,omitempty"`
    Res *vsantypes.Rename_TaskResponse `xml:"urn:vsan Rename_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *Rename_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func Rename_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.Rename_Task) (*vsantypes.Rename_TaskResponse, error) {
  var reqBody, resBody Rename_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReplaceCACertificatesAndCRLsBody struct{
    Req *vsantypes.ReplaceCACertificatesAndCRLs `xml:"urn:vsan ReplaceCACertificatesAndCRLs,omitempty"`
    Res *vsantypes.ReplaceCACertificatesAndCRLsResponse `xml:"urn:vsan ReplaceCACertificatesAndCRLsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReplaceCACertificatesAndCRLsBody) Fault() *soap.Fault { return b.Fault_ }

func ReplaceCACertificatesAndCRLs(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReplaceCACertificatesAndCRLs) (*vsantypes.ReplaceCACertificatesAndCRLsResponse, error) {
  var reqBody, resBody ReplaceCACertificatesAndCRLsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReplaceSmartCardTrustAnchorsBody struct{
    Req *vsantypes.ReplaceSmartCardTrustAnchors `xml:"urn:vsan ReplaceSmartCardTrustAnchors,omitempty"`
    Res *vsantypes.ReplaceSmartCardTrustAnchorsResponse `xml:"urn:vsan ReplaceSmartCardTrustAnchorsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReplaceSmartCardTrustAnchorsBody) Fault() *soap.Fault { return b.Fault_ }

func ReplaceSmartCardTrustAnchors(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReplaceSmartCardTrustAnchors) (*vsantypes.ReplaceSmartCardTrustAnchorsResponse, error) {
  var reqBody, resBody ReplaceSmartCardTrustAnchorsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RescanAllHbaBody struct{
    Req *vsantypes.RescanAllHba `xml:"urn:vsan RescanAllHba,omitempty"`
    Res *vsantypes.RescanAllHbaResponse `xml:"urn:vsan RescanAllHbaResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RescanAllHbaBody) Fault() *soap.Fault { return b.Fault_ }

func RescanAllHba(ctx context.Context, r soap.RoundTripper, req *vsantypes.RescanAllHba) (*vsantypes.RescanAllHbaResponse, error) {
  var reqBody, resBody RescanAllHbaBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RescanHbaBody struct{
    Req *vsantypes.RescanHba `xml:"urn:vsan RescanHba,omitempty"`
    Res *vsantypes.RescanHbaResponse `xml:"urn:vsan RescanHbaResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RescanHbaBody) Fault() *soap.Fault { return b.Fault_ }

func RescanHba(ctx context.Context, r soap.RoundTripper, req *vsantypes.RescanHba) (*vsantypes.RescanHbaResponse, error) {
  var reqBody, resBody RescanHbaBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RescanVffsBody struct{
    Req *vsantypes.RescanVffs `xml:"urn:vsan RescanVffs,omitempty"`
    Res *vsantypes.RescanVffsResponse `xml:"urn:vsan RescanVffsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RescanVffsBody) Fault() *soap.Fault { return b.Fault_ }

func RescanVffs(ctx context.Context, r soap.RoundTripper, req *vsantypes.RescanVffs) (*vsantypes.RescanVffsResponse, error) {
  var reqBody, resBody RescanVffsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RescanVmfsBody struct{
    Req *vsantypes.RescanVmfs `xml:"urn:vsan RescanVmfs,omitempty"`
    Res *vsantypes.RescanVmfsResponse `xml:"urn:vsan RescanVmfsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RescanVmfsBody) Fault() *soap.Fault { return b.Fault_ }

func RescanVmfs(ctx context.Context, r soap.RoundTripper, req *vsantypes.RescanVmfs) (*vsantypes.RescanVmfsResponse, error) {
  var reqBody, resBody RescanVmfsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetCollectorBody struct{
    Req *vsantypes.ResetCollector `xml:"urn:vsan ResetCollector,omitempty"`
    Res *vsantypes.ResetCollectorResponse `xml:"urn:vsan ResetCollectorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetCollectorBody) Fault() *soap.Fault { return b.Fault_ }

func ResetCollector(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetCollector) (*vsantypes.ResetCollectorResponse, error) {
  var reqBody, resBody ResetCollectorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetCounterLevelMappingBody struct{
    Req *vsantypes.ResetCounterLevelMapping `xml:"urn:vsan ResetCounterLevelMapping,omitempty"`
    Res *vsantypes.ResetCounterLevelMappingResponse `xml:"urn:vsan ResetCounterLevelMappingResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetCounterLevelMappingBody) Fault() *soap.Fault { return b.Fault_ }

func ResetCounterLevelMapping(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetCounterLevelMapping) (*vsantypes.ResetCounterLevelMappingResponse, error) {
  var reqBody, resBody ResetCounterLevelMappingBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetEntityPermissionsBody struct{
    Req *vsantypes.ResetEntityPermissions `xml:"urn:vsan ResetEntityPermissions,omitempty"`
    Res *vsantypes.ResetEntityPermissionsResponse `xml:"urn:vsan ResetEntityPermissionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetEntityPermissionsBody) Fault() *soap.Fault { return b.Fault_ }

func ResetEntityPermissions(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetEntityPermissions) (*vsantypes.ResetEntityPermissionsResponse, error) {
  var reqBody, resBody ResetEntityPermissionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetFirmwareToFactoryDefaultsBody struct{
    Req *vsantypes.ResetFirmwareToFactoryDefaults `xml:"urn:vsan ResetFirmwareToFactoryDefaults,omitempty"`
    Res *vsantypes.ResetFirmwareToFactoryDefaultsResponse `xml:"urn:vsan ResetFirmwareToFactoryDefaultsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetFirmwareToFactoryDefaultsBody) Fault() *soap.Fault { return b.Fault_ }

func ResetFirmwareToFactoryDefaults(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetFirmwareToFactoryDefaults) (*vsantypes.ResetFirmwareToFactoryDefaultsResponse, error) {
  var reqBody, resBody ResetFirmwareToFactoryDefaultsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetGuestInformationBody struct{
    Req *vsantypes.ResetGuestInformation `xml:"urn:vsan ResetGuestInformation,omitempty"`
    Res *vsantypes.ResetGuestInformationResponse `xml:"urn:vsan ResetGuestInformationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetGuestInformationBody) Fault() *soap.Fault { return b.Fault_ }

func ResetGuestInformation(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetGuestInformation) (*vsantypes.ResetGuestInformationResponse, error) {
  var reqBody, resBody ResetGuestInformationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetListViewBody struct{
    Req *vsantypes.ResetListView `xml:"urn:vsan ResetListView,omitempty"`
    Res *vsantypes.ResetListViewResponse `xml:"urn:vsan ResetListViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetListViewBody) Fault() *soap.Fault { return b.Fault_ }

func ResetListView(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetListView) (*vsantypes.ResetListViewResponse, error) {
  var reqBody, resBody ResetListViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetListViewFromViewBody struct{
    Req *vsantypes.ResetListViewFromView `xml:"urn:vsan ResetListViewFromView,omitempty"`
    Res *vsantypes.ResetListViewFromViewResponse `xml:"urn:vsan ResetListViewFromViewResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetListViewFromViewBody) Fault() *soap.Fault { return b.Fault_ }

func ResetListViewFromView(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetListViewFromView) (*vsantypes.ResetListViewFromViewResponse, error) {
  var reqBody, resBody ResetListViewFromViewBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetSystemHealthInfoBody struct{
    Req *vsantypes.ResetSystemHealthInfo `xml:"urn:vsan ResetSystemHealthInfo,omitempty"`
    Res *vsantypes.ResetSystemHealthInfoResponse `xml:"urn:vsan ResetSystemHealthInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetSystemHealthInfoBody) Fault() *soap.Fault { return b.Fault_ }

func ResetSystemHealthInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetSystemHealthInfo) (*vsantypes.ResetSystemHealthInfoResponse, error) {
  var reqBody, resBody ResetSystemHealthInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResetVM_TaskBody struct{
    Req *vsantypes.ResetVM_Task `xml:"urn:vsan ResetVM_Task,omitempty"`
    Res *vsantypes.ResetVM_TaskResponse `xml:"urn:vsan ResetVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResetVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ResetVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResetVM_Task) (*vsantypes.ResetVM_TaskResponse, error) {
  var reqBody, resBody ResetVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResignatureUnresolvedVmfsVolume_TaskBody struct{
    Req *vsantypes.ResignatureUnresolvedVmfsVolume_Task `xml:"urn:vsan ResignatureUnresolvedVmfsVolume_Task,omitempty"`
    Res *vsantypes.ResignatureUnresolvedVmfsVolume_TaskResponse `xml:"urn:vsan ResignatureUnresolvedVmfsVolume_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResignatureUnresolvedVmfsVolume_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ResignatureUnresolvedVmfsVolume_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResignatureUnresolvedVmfsVolume_Task) (*vsantypes.ResignatureUnresolvedVmfsVolume_TaskResponse, error) {
  var reqBody, resBody ResignatureUnresolvedVmfsVolume_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResolveInstallationErrorsOnCluster_TaskBody struct{
    Req *vsantypes.ResolveInstallationErrorsOnCluster_Task `xml:"urn:vsan ResolveInstallationErrorsOnCluster_Task,omitempty"`
    Res *vsantypes.ResolveInstallationErrorsOnCluster_TaskResponse `xml:"urn:vsan ResolveInstallationErrorsOnCluster_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResolveInstallationErrorsOnCluster_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ResolveInstallationErrorsOnCluster_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResolveInstallationErrorsOnCluster_Task) (*vsantypes.ResolveInstallationErrorsOnCluster_TaskResponse, error) {
  var reqBody, resBody ResolveInstallationErrorsOnCluster_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResolveInstallationErrorsOnHost_TaskBody struct{
    Req *vsantypes.ResolveInstallationErrorsOnHost_Task `xml:"urn:vsan ResolveInstallationErrorsOnHost_Task,omitempty"`
    Res *vsantypes.ResolveInstallationErrorsOnHost_TaskResponse `xml:"urn:vsan ResolveInstallationErrorsOnHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResolveInstallationErrorsOnHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ResolveInstallationErrorsOnHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResolveInstallationErrorsOnHost_Task) (*vsantypes.ResolveInstallationErrorsOnHost_TaskResponse, error) {
  var reqBody, resBody ResolveInstallationErrorsOnHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResolveMultipleUnresolvedVmfsVolumesBody struct{
    Req *vsantypes.ResolveMultipleUnresolvedVmfsVolumes `xml:"urn:vsan ResolveMultipleUnresolvedVmfsVolumes,omitempty"`
    Res *vsantypes.ResolveMultipleUnresolvedVmfsVolumesResponse `xml:"urn:vsan ResolveMultipleUnresolvedVmfsVolumesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResolveMultipleUnresolvedVmfsVolumesBody) Fault() *soap.Fault { return b.Fault_ }

func ResolveMultipleUnresolvedVmfsVolumes(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResolveMultipleUnresolvedVmfsVolumes) (*vsantypes.ResolveMultipleUnresolvedVmfsVolumesResponse, error) {
  var reqBody, resBody ResolveMultipleUnresolvedVmfsVolumesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ResolveMultipleUnresolvedVmfsVolumesEx_TaskBody struct{
    Req *vsantypes.ResolveMultipleUnresolvedVmfsVolumesEx_Task `xml:"urn:vsan ResolveMultipleUnresolvedVmfsVolumesEx_Task,omitempty"`
    Res *vsantypes.ResolveMultipleUnresolvedVmfsVolumesEx_TaskResponse `xml:"urn:vsan ResolveMultipleUnresolvedVmfsVolumesEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ResolveMultipleUnresolvedVmfsVolumesEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ResolveMultipleUnresolvedVmfsVolumesEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ResolveMultipleUnresolvedVmfsVolumesEx_Task) (*vsantypes.ResolveMultipleUnresolvedVmfsVolumesEx_TaskResponse, error) {
  var reqBody, resBody ResolveMultipleUnresolvedVmfsVolumesEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RestartServiceBody struct{
    Req *vsantypes.RestartService `xml:"urn:vsan RestartService,omitempty"`
    Res *vsantypes.RestartServiceResponse `xml:"urn:vsan RestartServiceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RestartServiceBody) Fault() *soap.Fault { return b.Fault_ }

func RestartService(ctx context.Context, r soap.RoundTripper, req *vsantypes.RestartService) (*vsantypes.RestartServiceResponse, error) {
  var reqBody, resBody RestartServiceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RestartServiceConsoleVirtualNicBody struct{
    Req *vsantypes.RestartServiceConsoleVirtualNic `xml:"urn:vsan RestartServiceConsoleVirtualNic,omitempty"`
    Res *vsantypes.RestartServiceConsoleVirtualNicResponse `xml:"urn:vsan RestartServiceConsoleVirtualNicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RestartServiceConsoleVirtualNicBody) Fault() *soap.Fault { return b.Fault_ }

func RestartServiceConsoleVirtualNic(ctx context.Context, r soap.RoundTripper, req *vsantypes.RestartServiceConsoleVirtualNic) (*vsantypes.RestartServiceConsoleVirtualNicResponse, error) {
  var reqBody, resBody RestartServiceConsoleVirtualNicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RestoreFirmwareConfigurationBody struct{
    Req *vsantypes.RestoreFirmwareConfiguration `xml:"urn:vsan RestoreFirmwareConfiguration,omitempty"`
    Res *vsantypes.RestoreFirmwareConfigurationResponse `xml:"urn:vsan RestoreFirmwareConfigurationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RestoreFirmwareConfigurationBody) Fault() *soap.Fault { return b.Fault_ }

func RestoreFirmwareConfiguration(ctx context.Context, r soap.RoundTripper, req *vsantypes.RestoreFirmwareConfiguration) (*vsantypes.RestoreFirmwareConfigurationResponse, error) {
  var reqBody, resBody RestoreFirmwareConfigurationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveAllFlashCapabilitiesBody struct{
    Req *vsantypes.RetrieveAllFlashCapabilities `xml:"urn:vsan RetrieveAllFlashCapabilities,omitempty"`
    Res *vsantypes.RetrieveAllFlashCapabilitiesResponse `xml:"urn:vsan RetrieveAllFlashCapabilitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveAllFlashCapabilitiesBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveAllFlashCapabilities(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveAllFlashCapabilities) (*vsantypes.RetrieveAllFlashCapabilitiesResponse, error) {
  var reqBody, resBody RetrieveAllFlashCapabilitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveAllPermissionsBody struct{
    Req *vsantypes.RetrieveAllPermissions `xml:"urn:vsan RetrieveAllPermissions,omitempty"`
    Res *vsantypes.RetrieveAllPermissionsResponse `xml:"urn:vsan RetrieveAllPermissionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveAllPermissionsBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveAllPermissions(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveAllPermissions) (*vsantypes.RetrieveAllPermissionsResponse, error) {
  var reqBody, resBody RetrieveAllPermissionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveAnswerFileBody struct{
    Req *vsantypes.RetrieveAnswerFile `xml:"urn:vsan RetrieveAnswerFile,omitempty"`
    Res *vsantypes.RetrieveAnswerFileResponse `xml:"urn:vsan RetrieveAnswerFileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveAnswerFileBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveAnswerFile(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveAnswerFile) (*vsantypes.RetrieveAnswerFileResponse, error) {
  var reqBody, resBody RetrieveAnswerFileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveAnswerFileForProfileBody struct{
    Req *vsantypes.RetrieveAnswerFileForProfile `xml:"urn:vsan RetrieveAnswerFileForProfile,omitempty"`
    Res *vsantypes.RetrieveAnswerFileForProfileResponse `xml:"urn:vsan RetrieveAnswerFileForProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveAnswerFileForProfileBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveAnswerFileForProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveAnswerFileForProfile) (*vsantypes.RetrieveAnswerFileForProfileResponse, error) {
  var reqBody, resBody RetrieveAnswerFileForProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveArgumentDescriptionBody struct{
    Req *vsantypes.RetrieveArgumentDescription `xml:"urn:vsan RetrieveArgumentDescription,omitempty"`
    Res *vsantypes.RetrieveArgumentDescriptionResponse `xml:"urn:vsan RetrieveArgumentDescriptionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveArgumentDescriptionBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveArgumentDescription(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveArgumentDescription) (*vsantypes.RetrieveArgumentDescriptionResponse, error) {
  var reqBody, resBody RetrieveArgumentDescriptionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveClientCertBody struct{
    Req *vsantypes.RetrieveClientCert `xml:"urn:vsan RetrieveClientCert,omitempty"`
    Res *vsantypes.RetrieveClientCertResponse `xml:"urn:vsan RetrieveClientCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveClientCertBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveClientCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveClientCert) (*vsantypes.RetrieveClientCertResponse, error) {
  var reqBody, resBody RetrieveClientCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveClientCsrBody struct{
    Req *vsantypes.RetrieveClientCsr `xml:"urn:vsan RetrieveClientCsr,omitempty"`
    Res *vsantypes.RetrieveClientCsrResponse `xml:"urn:vsan RetrieveClientCsrResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveClientCsrBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveClientCsr(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveClientCsr) (*vsantypes.RetrieveClientCsrResponse, error) {
  var reqBody, resBody RetrieveClientCsrBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveDasAdvancedRuntimeInfoBody struct{
    Req *vsantypes.RetrieveDasAdvancedRuntimeInfo `xml:"urn:vsan RetrieveDasAdvancedRuntimeInfo,omitempty"`
    Res *vsantypes.RetrieveDasAdvancedRuntimeInfoResponse `xml:"urn:vsan RetrieveDasAdvancedRuntimeInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveDasAdvancedRuntimeInfoBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveDasAdvancedRuntimeInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveDasAdvancedRuntimeInfo) (*vsantypes.RetrieveDasAdvancedRuntimeInfoResponse, error) {
  var reqBody, resBody RetrieveDasAdvancedRuntimeInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveDescriptionBody struct{
    Req *vsantypes.RetrieveDescription `xml:"urn:vsan RetrieveDescription,omitempty"`
    Res *vsantypes.RetrieveDescriptionResponse `xml:"urn:vsan RetrieveDescriptionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveDescriptionBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveDescription(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveDescription) (*vsantypes.RetrieveDescriptionResponse, error) {
  var reqBody, resBody RetrieveDescriptionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveDiskPartitionInfoBody struct{
    Req *vsantypes.RetrieveDiskPartitionInfo `xml:"urn:vsan RetrieveDiskPartitionInfo,omitempty"`
    Res *vsantypes.RetrieveDiskPartitionInfoResponse `xml:"urn:vsan RetrieveDiskPartitionInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveDiskPartitionInfoBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveDiskPartitionInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveDiskPartitionInfo) (*vsantypes.RetrieveDiskPartitionInfoResponse, error) {
  var reqBody, resBody RetrieveDiskPartitionInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveEntityPermissionsBody struct{
    Req *vsantypes.RetrieveEntityPermissions `xml:"urn:vsan RetrieveEntityPermissions,omitempty"`
    Res *vsantypes.RetrieveEntityPermissionsResponse `xml:"urn:vsan RetrieveEntityPermissionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveEntityPermissionsBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveEntityPermissions(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveEntityPermissions) (*vsantypes.RetrieveEntityPermissionsResponse, error) {
  var reqBody, resBody RetrieveEntityPermissionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveEntityScheduledTaskBody struct{
    Req *vsantypes.RetrieveEntityScheduledTask `xml:"urn:vsan RetrieveEntityScheduledTask,omitempty"`
    Res *vsantypes.RetrieveEntityScheduledTaskResponse `xml:"urn:vsan RetrieveEntityScheduledTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveEntityScheduledTaskBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveEntityScheduledTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveEntityScheduledTask) (*vsantypes.RetrieveEntityScheduledTaskResponse, error) {
  var reqBody, resBody RetrieveEntityScheduledTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveHardwareUptimeBody struct{
    Req *vsantypes.RetrieveHardwareUptime `xml:"urn:vsan RetrieveHardwareUptime,omitempty"`
    Res *vsantypes.RetrieveHardwareUptimeResponse `xml:"urn:vsan RetrieveHardwareUptimeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveHardwareUptimeBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveHardwareUptime(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveHardwareUptime) (*vsantypes.RetrieveHardwareUptimeResponse, error) {
  var reqBody, resBody RetrieveHardwareUptimeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveHostAccessControlEntriesBody struct{
    Req *vsantypes.RetrieveHostAccessControlEntries `xml:"urn:vsan RetrieveHostAccessControlEntries,omitempty"`
    Res *vsantypes.RetrieveHostAccessControlEntriesResponse `xml:"urn:vsan RetrieveHostAccessControlEntriesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveHostAccessControlEntriesBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveHostAccessControlEntries(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveHostAccessControlEntries) (*vsantypes.RetrieveHostAccessControlEntriesResponse, error) {
  var reqBody, resBody RetrieveHostAccessControlEntriesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveHostCustomizationsBody struct{
    Req *vsantypes.RetrieveHostCustomizations `xml:"urn:vsan RetrieveHostCustomizations,omitempty"`
    Res *vsantypes.RetrieveHostCustomizationsResponse `xml:"urn:vsan RetrieveHostCustomizationsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveHostCustomizationsBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveHostCustomizations(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveHostCustomizations) (*vsantypes.RetrieveHostCustomizationsResponse, error) {
  var reqBody, resBody RetrieveHostCustomizationsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveHostCustomizationsForProfileBody struct{
    Req *vsantypes.RetrieveHostCustomizationsForProfile `xml:"urn:vsan RetrieveHostCustomizationsForProfile,omitempty"`
    Res *vsantypes.RetrieveHostCustomizationsForProfileResponse `xml:"urn:vsan RetrieveHostCustomizationsForProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveHostCustomizationsForProfileBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveHostCustomizationsForProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveHostCustomizationsForProfile) (*vsantypes.RetrieveHostCustomizationsForProfileResponse, error) {
  var reqBody, resBody RetrieveHostCustomizationsForProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveHostSpecificationBody struct{
    Req *vsantypes.RetrieveHostSpecification `xml:"urn:vsan RetrieveHostSpecification,omitempty"`
    Res *vsantypes.RetrieveHostSpecificationResponse `xml:"urn:vsan RetrieveHostSpecificationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveHostSpecificationBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveHostSpecification(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveHostSpecification) (*vsantypes.RetrieveHostSpecificationResponse, error) {
  var reqBody, resBody RetrieveHostSpecificationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveKmipServerCertBody struct{
    Req *vsantypes.RetrieveKmipServerCert `xml:"urn:vsan RetrieveKmipServerCert,omitempty"`
    Res *vsantypes.RetrieveKmipServerCertResponse `xml:"urn:vsan RetrieveKmipServerCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveKmipServerCertBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveKmipServerCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveKmipServerCert) (*vsantypes.RetrieveKmipServerCertResponse, error) {
  var reqBody, resBody RetrieveKmipServerCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveKmipServersStatus_TaskBody struct{
    Req *vsantypes.RetrieveKmipServersStatus_Task `xml:"urn:vsan RetrieveKmipServersStatus_Task,omitempty"`
    Res *vsantypes.RetrieveKmipServersStatus_TaskResponse `xml:"urn:vsan RetrieveKmipServersStatus_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveKmipServersStatus_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveKmipServersStatus_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveKmipServersStatus_Task) (*vsantypes.RetrieveKmipServersStatus_TaskResponse, error) {
  var reqBody, resBody RetrieveKmipServersStatus_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveObjectScheduledTaskBody struct{
    Req *vsantypes.RetrieveObjectScheduledTask `xml:"urn:vsan RetrieveObjectScheduledTask,omitempty"`
    Res *vsantypes.RetrieveObjectScheduledTaskResponse `xml:"urn:vsan RetrieveObjectScheduledTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveObjectScheduledTaskBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveObjectScheduledTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveObjectScheduledTask) (*vsantypes.RetrieveObjectScheduledTaskResponse, error) {
  var reqBody, resBody RetrieveObjectScheduledTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveProductComponentsBody struct{
    Req *vsantypes.RetrieveProductComponents `xml:"urn:vsan RetrieveProductComponents,omitempty"`
    Res *vsantypes.RetrieveProductComponentsResponse `xml:"urn:vsan RetrieveProductComponentsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveProductComponentsBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveProductComponents(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveProductComponents) (*vsantypes.RetrieveProductComponentsResponse, error) {
  var reqBody, resBody RetrieveProductComponentsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrievePropertiesBody struct{
    Req *vsantypes.RetrieveProperties `xml:"urn:vsan RetrieveProperties,omitempty"`
    Res *vsantypes.RetrievePropertiesResponse `xml:"urn:vsan RetrievePropertiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrievePropertiesBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveProperties(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveProperties) (*vsantypes.RetrievePropertiesResponse, error) {
  var reqBody, resBody RetrievePropertiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrievePropertiesExBody struct{
    Req *vsantypes.RetrievePropertiesEx `xml:"urn:vsan RetrievePropertiesEx,omitempty"`
    Res *vsantypes.RetrievePropertiesExResponse `xml:"urn:vsan RetrievePropertiesExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrievePropertiesExBody) Fault() *soap.Fault { return b.Fault_ }

func RetrievePropertiesEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrievePropertiesEx) (*vsantypes.RetrievePropertiesExResponse, error) {
  var reqBody, resBody RetrievePropertiesExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveRolePermissionsBody struct{
    Req *vsantypes.RetrieveRolePermissions `xml:"urn:vsan RetrieveRolePermissions,omitempty"`
    Res *vsantypes.RetrieveRolePermissionsResponse `xml:"urn:vsan RetrieveRolePermissionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveRolePermissionsBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveRolePermissions(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveRolePermissions) (*vsantypes.RetrieveRolePermissionsResponse, error) {
  var reqBody, resBody RetrieveRolePermissionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveSelfSignedClientCertBody struct{
    Req *vsantypes.RetrieveSelfSignedClientCert `xml:"urn:vsan RetrieveSelfSignedClientCert,omitempty"`
    Res *vsantypes.RetrieveSelfSignedClientCertResponse `xml:"urn:vsan RetrieveSelfSignedClientCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveSelfSignedClientCertBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveSelfSignedClientCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveSelfSignedClientCert) (*vsantypes.RetrieveSelfSignedClientCertResponse, error) {
  var reqBody, resBody RetrieveSelfSignedClientCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveServiceContentBody struct{
    Req *vsantypes.RetrieveServiceContent `xml:"urn:vsan RetrieveServiceContent,omitempty"`
    Res *vsantypes.RetrieveServiceContentResponse `xml:"urn:vsan RetrieveServiceContentResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveServiceContentBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveServiceContent(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveServiceContent) (*vsantypes.RetrieveServiceContentResponse, error) {
  var reqBody, resBody RetrieveServiceContentBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveSupportedVsanFormatVersionBody struct{
    Req *vsantypes.RetrieveSupportedVsanFormatVersion `xml:"urn:vsan RetrieveSupportedVsanFormatVersion,omitempty"`
    Res *vsantypes.RetrieveSupportedVsanFormatVersionResponse `xml:"urn:vsan RetrieveSupportedVsanFormatVersionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveSupportedVsanFormatVersionBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveSupportedVsanFormatVersion(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveSupportedVsanFormatVersion) (*vsantypes.RetrieveSupportedVsanFormatVersionResponse, error) {
  var reqBody, resBody RetrieveSupportedVsanFormatVersionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveUserGroupsBody struct{
    Req *vsantypes.RetrieveUserGroups `xml:"urn:vsan RetrieveUserGroups,omitempty"`
    Res *vsantypes.RetrieveUserGroupsResponse `xml:"urn:vsan RetrieveUserGroupsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveUserGroupsBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveUserGroups(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveUserGroups) (*vsantypes.RetrieveUserGroupsResponse, error) {
  var reqBody, resBody RetrieveUserGroupsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveVStorageObjectBody struct{
    Req *vsantypes.RetrieveVStorageObject `xml:"urn:vsan RetrieveVStorageObject,omitempty"`
    Res *vsantypes.RetrieveVStorageObjectResponse `xml:"urn:vsan RetrieveVStorageObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveVStorageObjectBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveVStorageObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveVStorageObject) (*vsantypes.RetrieveVStorageObjectResponse, error) {
  var reqBody, resBody RetrieveVStorageObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RetrieveVStorageObjectStateBody struct{
    Req *vsantypes.RetrieveVStorageObjectState `xml:"urn:vsan RetrieveVStorageObjectState,omitempty"`
    Res *vsantypes.RetrieveVStorageObjectStateResponse `xml:"urn:vsan RetrieveVStorageObjectStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RetrieveVStorageObjectStateBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveVStorageObjectState(ctx context.Context, r soap.RoundTripper, req *vsantypes.RetrieveVStorageObjectState) (*vsantypes.RetrieveVStorageObjectStateResponse, error) {
  var reqBody, resBody RetrieveVStorageObjectStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RevertToCurrentSnapshot_TaskBody struct{
    Req *vsantypes.RevertToCurrentSnapshot_Task `xml:"urn:vsan RevertToCurrentSnapshot_Task,omitempty"`
    Res *vsantypes.RevertToCurrentSnapshot_TaskResponse `xml:"urn:vsan RevertToCurrentSnapshot_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RevertToCurrentSnapshot_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RevertToCurrentSnapshot_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RevertToCurrentSnapshot_Task) (*vsantypes.RevertToCurrentSnapshot_TaskResponse, error) {
  var reqBody, resBody RevertToCurrentSnapshot_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RevertToSnapshot_TaskBody struct{
    Req *vsantypes.RevertToSnapshot_Task `xml:"urn:vsan RevertToSnapshot_Task,omitempty"`
    Res *vsantypes.RevertToSnapshot_TaskResponse `xml:"urn:vsan RevertToSnapshot_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RevertToSnapshot_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func RevertToSnapshot_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.RevertToSnapshot_Task) (*vsantypes.RevertToSnapshot_TaskResponse, error) {
  var reqBody, resBody RevertToSnapshot_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RewindCollectorBody struct{
    Req *vsantypes.RewindCollector `xml:"urn:vsan RewindCollector,omitempty"`
    Res *vsantypes.RewindCollectorResponse `xml:"urn:vsan RewindCollectorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RewindCollectorBody) Fault() *soap.Fault { return b.Fault_ }

func RewindCollector(ctx context.Context, r soap.RoundTripper, req *vsantypes.RewindCollector) (*vsantypes.RewindCollectorResponse, error) {
  var reqBody, resBody RewindCollectorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RunScheduledTaskBody struct{
    Req *vsantypes.RunScheduledTask `xml:"urn:vsan RunScheduledTask,omitempty"`
    Res *vsantypes.RunScheduledTaskResponse `xml:"urn:vsan RunScheduledTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RunScheduledTaskBody) Fault() *soap.Fault { return b.Fault_ }

func RunScheduledTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.RunScheduledTask) (*vsantypes.RunScheduledTaskResponse, error) {
  var reqBody, resBody RunScheduledTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type RunVsanPhysicalDiskDiagnosticsBody struct{
    Req *vsantypes.RunVsanPhysicalDiskDiagnostics `xml:"urn:vsan RunVsanPhysicalDiskDiagnostics,omitempty"`
    Res *vsantypes.RunVsanPhysicalDiskDiagnosticsResponse `xml:"urn:vsan RunVsanPhysicalDiskDiagnosticsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *RunVsanPhysicalDiskDiagnosticsBody) Fault() *soap.Fault { return b.Fault_ }

func RunVsanPhysicalDiskDiagnostics(ctx context.Context, r soap.RoundTripper, req *vsantypes.RunVsanPhysicalDiskDiagnostics) (*vsantypes.RunVsanPhysicalDiskDiagnosticsResponse, error) {
  var reqBody, resBody RunVsanPhysicalDiskDiagnosticsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ScanHostPatchV2_TaskBody struct{
    Req *vsantypes.ScanHostPatchV2_Task `xml:"urn:vsan ScanHostPatchV2_Task,omitempty"`
    Res *vsantypes.ScanHostPatchV2_TaskResponse `xml:"urn:vsan ScanHostPatchV2_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ScanHostPatchV2_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ScanHostPatchV2_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ScanHostPatchV2_Task) (*vsantypes.ScanHostPatchV2_TaskResponse, error) {
  var reqBody, resBody ScanHostPatchV2_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ScanHostPatch_TaskBody struct{
    Req *vsantypes.ScanHostPatch_Task `xml:"urn:vsan ScanHostPatch_Task,omitempty"`
    Res *vsantypes.ScanHostPatch_TaskResponse `xml:"urn:vsan ScanHostPatch_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ScanHostPatch_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ScanHostPatch_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ScanHostPatch_Task) (*vsantypes.ScanHostPatch_TaskResponse, error) {
  var reqBody, resBody ScanHostPatch_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ScheduleReconcileDatastoreInventoryBody struct{
    Req *vsantypes.ScheduleReconcileDatastoreInventory `xml:"urn:vsan ScheduleReconcileDatastoreInventory,omitempty"`
    Res *vsantypes.ScheduleReconcileDatastoreInventoryResponse `xml:"urn:vsan ScheduleReconcileDatastoreInventoryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ScheduleReconcileDatastoreInventoryBody) Fault() *soap.Fault { return b.Fault_ }

func ScheduleReconcileDatastoreInventory(ctx context.Context, r soap.RoundTripper, req *vsantypes.ScheduleReconcileDatastoreInventory) (*vsantypes.ScheduleReconcileDatastoreInventoryResponse, error) {
  var reqBody, resBody ScheduleReconcileDatastoreInventoryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SearchDatastoreSubFolders_TaskBody struct{
    Req *vsantypes.SearchDatastoreSubFolders_Task `xml:"urn:vsan SearchDatastoreSubFolders_Task,omitempty"`
    Res *vsantypes.SearchDatastoreSubFolders_TaskResponse `xml:"urn:vsan SearchDatastoreSubFolders_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SearchDatastoreSubFolders_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func SearchDatastoreSubFolders_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.SearchDatastoreSubFolders_Task) (*vsantypes.SearchDatastoreSubFolders_TaskResponse, error) {
  var reqBody, resBody SearchDatastoreSubFolders_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SearchDatastore_TaskBody struct{
    Req *vsantypes.SearchDatastore_Task `xml:"urn:vsan SearchDatastore_Task,omitempty"`
    Res *vsantypes.SearchDatastore_TaskResponse `xml:"urn:vsan SearchDatastore_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SearchDatastore_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func SearchDatastore_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.SearchDatastore_Task) (*vsantypes.SearchDatastore_TaskResponse, error) {
  var reqBody, resBody SearchDatastore_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SelectActivePartitionBody struct{
    Req *vsantypes.SelectActivePartition `xml:"urn:vsan SelectActivePartition,omitempty"`
    Res *vsantypes.SelectActivePartitionResponse `xml:"urn:vsan SelectActivePartitionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SelectActivePartitionBody) Fault() *soap.Fault { return b.Fault_ }

func SelectActivePartition(ctx context.Context, r soap.RoundTripper, req *vsantypes.SelectActivePartition) (*vsantypes.SelectActivePartitionResponse, error) {
  var reqBody, resBody SelectActivePartitionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SelectVnicBody struct{
    Req *vsantypes.SelectVnic `xml:"urn:vsan SelectVnic,omitempty"`
    Res *vsantypes.SelectVnicResponse `xml:"urn:vsan SelectVnicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SelectVnicBody) Fault() *soap.Fault { return b.Fault_ }

func SelectVnic(ctx context.Context, r soap.RoundTripper, req *vsantypes.SelectVnic) (*vsantypes.SelectVnicResponse, error) {
  var reqBody, resBody SelectVnicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SelectVnicForNicTypeBody struct{
    Req *vsantypes.SelectVnicForNicType `xml:"urn:vsan SelectVnicForNicType,omitempty"`
    Res *vsantypes.SelectVnicForNicTypeResponse `xml:"urn:vsan SelectVnicForNicTypeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SelectVnicForNicTypeBody) Fault() *soap.Fault { return b.Fault_ }

func SelectVnicForNicType(ctx context.Context, r soap.RoundTripper, req *vsantypes.SelectVnicForNicType) (*vsantypes.SelectVnicForNicTypeResponse, error) {
  var reqBody, resBody SelectVnicForNicTypeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SendNMIBody struct{
    Req *vsantypes.SendNMI `xml:"urn:vsan SendNMI,omitempty"`
    Res *vsantypes.SendNMIResponse `xml:"urn:vsan SendNMIResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SendNMIBody) Fault() *soap.Fault { return b.Fault_ }

func SendNMI(ctx context.Context, r soap.RoundTripper, req *vsantypes.SendNMI) (*vsantypes.SendNMIResponse, error) {
  var reqBody, resBody SendNMIBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SendTestNotificationBody struct{
    Req *vsantypes.SendTestNotification `xml:"urn:vsan SendTestNotification,omitempty"`
    Res *vsantypes.SendTestNotificationResponse `xml:"urn:vsan SendTestNotificationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SendTestNotificationBody) Fault() *soap.Fault { return b.Fault_ }

func SendTestNotification(ctx context.Context, r soap.RoundTripper, req *vsantypes.SendTestNotification) (*vsantypes.SendTestNotificationResponse, error) {
  var reqBody, resBody SendTestNotificationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SessionIsActiveBody struct{
    Req *vsantypes.SessionIsActive `xml:"urn:vsan SessionIsActive,omitempty"`
    Res *vsantypes.SessionIsActiveResponse `xml:"urn:vsan SessionIsActiveResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SessionIsActiveBody) Fault() *soap.Fault { return b.Fault_ }

func SessionIsActive(ctx context.Context, r soap.RoundTripper, req *vsantypes.SessionIsActive) (*vsantypes.SessionIsActiveResponse, error) {
  var reqBody, resBody SessionIsActiveBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetCollectorPageSizeBody struct{
    Req *vsantypes.SetCollectorPageSize `xml:"urn:vsan SetCollectorPageSize,omitempty"`
    Res *vsantypes.SetCollectorPageSizeResponse `xml:"urn:vsan SetCollectorPageSizeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetCollectorPageSizeBody) Fault() *soap.Fault { return b.Fault_ }

func SetCollectorPageSize(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetCollectorPageSize) (*vsantypes.SetCollectorPageSizeResponse, error) {
  var reqBody, resBody SetCollectorPageSizeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetDisplayTopologyBody struct{
    Req *vsantypes.SetDisplayTopology `xml:"urn:vsan SetDisplayTopology,omitempty"`
    Res *vsantypes.SetDisplayTopologyResponse `xml:"urn:vsan SetDisplayTopologyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetDisplayTopologyBody) Fault() *soap.Fault { return b.Fault_ }

func SetDisplayTopology(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetDisplayTopology) (*vsantypes.SetDisplayTopologyResponse, error) {
  var reqBody, resBody SetDisplayTopologyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetEntityPermissionsBody struct{
    Req *vsantypes.SetEntityPermissions `xml:"urn:vsan SetEntityPermissions,omitempty"`
    Res *vsantypes.SetEntityPermissionsResponse `xml:"urn:vsan SetEntityPermissionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetEntityPermissionsBody) Fault() *soap.Fault { return b.Fault_ }

func SetEntityPermissions(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetEntityPermissions) (*vsantypes.SetEntityPermissionsResponse, error) {
  var reqBody, resBody SetEntityPermissionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetExtensionCertificateBody struct{
    Req *vsantypes.SetExtensionCertificate `xml:"urn:vsan SetExtensionCertificate,omitempty"`
    Res *vsantypes.SetExtensionCertificateResponse `xml:"urn:vsan SetExtensionCertificateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetExtensionCertificateBody) Fault() *soap.Fault { return b.Fault_ }

func SetExtensionCertificate(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetExtensionCertificate) (*vsantypes.SetExtensionCertificateResponse, error) {
  var reqBody, resBody SetExtensionCertificateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetFieldBody struct{
    Req *vsantypes.SetField `xml:"urn:vsan SetField,omitempty"`
    Res *vsantypes.SetFieldResponse `xml:"urn:vsan SetFieldResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetFieldBody) Fault() *soap.Fault { return b.Fault_ }

func SetField(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetField) (*vsantypes.SetFieldResponse, error) {
  var reqBody, resBody SetFieldBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetLicenseEditionBody struct{
    Req *vsantypes.SetLicenseEdition `xml:"urn:vsan SetLicenseEdition,omitempty"`
    Res *vsantypes.SetLicenseEditionResponse `xml:"urn:vsan SetLicenseEditionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetLicenseEditionBody) Fault() *soap.Fault { return b.Fault_ }

func SetLicenseEdition(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetLicenseEdition) (*vsantypes.SetLicenseEditionResponse, error) {
  var reqBody, resBody SetLicenseEditionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetLocaleBody struct{
    Req *vsantypes.SetLocale `xml:"urn:vsan SetLocale,omitempty"`
    Res *vsantypes.SetLocaleResponse `xml:"urn:vsan SetLocaleResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetLocaleBody) Fault() *soap.Fault { return b.Fault_ }

func SetLocale(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetLocale) (*vsantypes.SetLocaleResponse, error) {
  var reqBody, resBody SetLocaleBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetMultipathLunPolicyBody struct{
    Req *vsantypes.SetMultipathLunPolicy `xml:"urn:vsan SetMultipathLunPolicy,omitempty"`
    Res *vsantypes.SetMultipathLunPolicyResponse `xml:"urn:vsan SetMultipathLunPolicyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetMultipathLunPolicyBody) Fault() *soap.Fault { return b.Fault_ }

func SetMultipathLunPolicy(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetMultipathLunPolicy) (*vsantypes.SetMultipathLunPolicyResponse, error) {
  var reqBody, resBody SetMultipathLunPolicyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetNFSUserBody struct{
    Req *vsantypes.SetNFSUser `xml:"urn:vsan SetNFSUser,omitempty"`
    Res *vsantypes.SetNFSUserResponse `xml:"urn:vsan SetNFSUserResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetNFSUserBody) Fault() *soap.Fault { return b.Fault_ }

func SetNFSUser(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetNFSUser) (*vsantypes.SetNFSUserResponse, error) {
  var reqBody, resBody SetNFSUserBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetPublicKeyBody struct{
    Req *vsantypes.SetPublicKey `xml:"urn:vsan SetPublicKey,omitempty"`
    Res *vsantypes.SetPublicKeyResponse `xml:"urn:vsan SetPublicKeyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetPublicKeyBody) Fault() *soap.Fault { return b.Fault_ }

func SetPublicKey(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetPublicKey) (*vsantypes.SetPublicKeyResponse, error) {
  var reqBody, resBody SetPublicKeyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetRegistryValueInGuestBody struct{
    Req *vsantypes.SetRegistryValueInGuest `xml:"urn:vsan SetRegistryValueInGuest,omitempty"`
    Res *vsantypes.SetRegistryValueInGuestResponse `xml:"urn:vsan SetRegistryValueInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetRegistryValueInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func SetRegistryValueInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetRegistryValueInGuest) (*vsantypes.SetRegistryValueInGuestResponse, error) {
  var reqBody, resBody SetRegistryValueInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetScreenResolutionBody struct{
    Req *vsantypes.SetScreenResolution `xml:"urn:vsan SetScreenResolution,omitempty"`
    Res *vsantypes.SetScreenResolutionResponse `xml:"urn:vsan SetScreenResolutionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetScreenResolutionBody) Fault() *soap.Fault { return b.Fault_ }

func SetScreenResolution(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetScreenResolution) (*vsantypes.SetScreenResolutionResponse, error) {
  var reqBody, resBody SetScreenResolutionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetTaskDescriptionBody struct{
    Req *vsantypes.SetTaskDescription `xml:"urn:vsan SetTaskDescription,omitempty"`
    Res *vsantypes.SetTaskDescriptionResponse `xml:"urn:vsan SetTaskDescriptionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetTaskDescriptionBody) Fault() *soap.Fault { return b.Fault_ }

func SetTaskDescription(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetTaskDescription) (*vsantypes.SetTaskDescriptionResponse, error) {
  var reqBody, resBody SetTaskDescriptionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetTaskStateBody struct{
    Req *vsantypes.SetTaskState `xml:"urn:vsan SetTaskState,omitempty"`
    Res *vsantypes.SetTaskStateResponse `xml:"urn:vsan SetTaskStateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetTaskStateBody) Fault() *soap.Fault { return b.Fault_ }

func SetTaskState(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetTaskState) (*vsantypes.SetTaskStateResponse, error) {
  var reqBody, resBody SetTaskStateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetVirtualDiskUuidBody struct{
    Req *vsantypes.SetVirtualDiskUuid `xml:"urn:vsan SetVirtualDiskUuid,omitempty"`
    Res *vsantypes.SetVirtualDiskUuidResponse `xml:"urn:vsan SetVirtualDiskUuidResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetVirtualDiskUuidBody) Fault() *soap.Fault { return b.Fault_ }

func SetVirtualDiskUuid(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetVirtualDiskUuid) (*vsantypes.SetVirtualDiskUuidResponse, error) {
  var reqBody, resBody SetVirtualDiskUuidBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ShrinkVirtualDisk_TaskBody struct{
    Req *vsantypes.ShrinkVirtualDisk_Task `xml:"urn:vsan ShrinkVirtualDisk_Task,omitempty"`
    Res *vsantypes.ShrinkVirtualDisk_TaskResponse `xml:"urn:vsan ShrinkVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ShrinkVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ShrinkVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ShrinkVirtualDisk_Task) (*vsantypes.ShrinkVirtualDisk_TaskResponse, error) {
  var reqBody, resBody ShrinkVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ShutdownGuestBody struct{
    Req *vsantypes.ShutdownGuest `xml:"urn:vsan ShutdownGuest,omitempty"`
    Res *vsantypes.ShutdownGuestResponse `xml:"urn:vsan ShutdownGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ShutdownGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ShutdownGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ShutdownGuest) (*vsantypes.ShutdownGuestResponse, error) {
  var reqBody, resBody ShutdownGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ShutdownHost_TaskBody struct{
    Req *vsantypes.ShutdownHost_Task `xml:"urn:vsan ShutdownHost_Task,omitempty"`
    Res *vsantypes.ShutdownHost_TaskResponse `xml:"urn:vsan ShutdownHost_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ShutdownHost_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ShutdownHost_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ShutdownHost_Task) (*vsantypes.ShutdownHost_TaskResponse, error) {
  var reqBody, resBody ShutdownHost_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StageHostPatch_TaskBody struct{
    Req *vsantypes.StageHostPatch_Task `xml:"urn:vsan StageHostPatch_Task,omitempty"`
    Res *vsantypes.StageHostPatch_TaskResponse `xml:"urn:vsan StageHostPatch_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StageHostPatch_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func StageHostPatch_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.StageHostPatch_Task) (*vsantypes.StageHostPatch_TaskResponse, error) {
  var reqBody, resBody StageHostPatch_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StampAllRulesWithUuid_TaskBody struct{
    Req *vsantypes.StampAllRulesWithUuid_Task `xml:"urn:vsan StampAllRulesWithUuid_Task,omitempty"`
    Res *vsantypes.StampAllRulesWithUuid_TaskResponse `xml:"urn:vsan StampAllRulesWithUuid_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StampAllRulesWithUuid_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func StampAllRulesWithUuid_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.StampAllRulesWithUuid_Task) (*vsantypes.StampAllRulesWithUuid_TaskResponse, error) {
  var reqBody, resBody StampAllRulesWithUuid_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StandbyGuestBody struct{
    Req *vsantypes.StandbyGuest `xml:"urn:vsan StandbyGuest,omitempty"`
    Res *vsantypes.StandbyGuestResponse `xml:"urn:vsan StandbyGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StandbyGuestBody) Fault() *soap.Fault { return b.Fault_ }

func StandbyGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.StandbyGuest) (*vsantypes.StandbyGuestResponse, error) {
  var reqBody, resBody StandbyGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StartProgramInGuestBody struct{
    Req *vsantypes.StartProgramInGuest `xml:"urn:vsan StartProgramInGuest,omitempty"`
    Res *vsantypes.StartProgramInGuestResponse `xml:"urn:vsan StartProgramInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StartProgramInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func StartProgramInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.StartProgramInGuest) (*vsantypes.StartProgramInGuestResponse, error) {
  var reqBody, resBody StartProgramInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StartRecording_TaskBody struct{
    Req *vsantypes.StartRecording_Task `xml:"urn:vsan StartRecording_Task,omitempty"`
    Res *vsantypes.StartRecording_TaskResponse `xml:"urn:vsan StartRecording_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StartRecording_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func StartRecording_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.StartRecording_Task) (*vsantypes.StartRecording_TaskResponse, error) {
  var reqBody, resBody StartRecording_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StartReplaying_TaskBody struct{
    Req *vsantypes.StartReplaying_Task `xml:"urn:vsan StartReplaying_Task,omitempty"`
    Res *vsantypes.StartReplaying_TaskResponse `xml:"urn:vsan StartReplaying_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StartReplaying_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func StartReplaying_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.StartReplaying_Task) (*vsantypes.StartReplaying_TaskResponse, error) {
  var reqBody, resBody StartReplaying_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StartServiceBody struct{
    Req *vsantypes.StartService `xml:"urn:vsan StartService,omitempty"`
    Res *vsantypes.StartServiceResponse `xml:"urn:vsan StartServiceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StartServiceBody) Fault() *soap.Fault { return b.Fault_ }

func StartService(ctx context.Context, r soap.RoundTripper, req *vsantypes.StartService) (*vsantypes.StartServiceResponse, error) {
  var reqBody, resBody StartServiceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StopRecording_TaskBody struct{
    Req *vsantypes.StopRecording_Task `xml:"urn:vsan StopRecording_Task,omitempty"`
    Res *vsantypes.StopRecording_TaskResponse `xml:"urn:vsan StopRecording_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StopRecording_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func StopRecording_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.StopRecording_Task) (*vsantypes.StopRecording_TaskResponse, error) {
  var reqBody, resBody StopRecording_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StopReplaying_TaskBody struct{
    Req *vsantypes.StopReplaying_Task `xml:"urn:vsan StopReplaying_Task,omitempty"`
    Res *vsantypes.StopReplaying_TaskResponse `xml:"urn:vsan StopReplaying_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StopReplaying_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func StopReplaying_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.StopReplaying_Task) (*vsantypes.StopReplaying_TaskResponse, error) {
  var reqBody, resBody StopReplaying_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type StopServiceBody struct{
    Req *vsantypes.StopService `xml:"urn:vsan StopService,omitempty"`
    Res *vsantypes.StopServiceResponse `xml:"urn:vsan StopServiceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *StopServiceBody) Fault() *soap.Fault { return b.Fault_ }

func StopService(ctx context.Context, r soap.RoundTripper, req *vsantypes.StopService) (*vsantypes.StopServiceResponse, error) {
  var reqBody, resBody StopServiceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SuspendVApp_TaskBody struct{
    Req *vsantypes.SuspendVApp_Task `xml:"urn:vsan SuspendVApp_Task,omitempty"`
    Res *vsantypes.SuspendVApp_TaskResponse `xml:"urn:vsan SuspendVApp_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SuspendVApp_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func SuspendVApp_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.SuspendVApp_Task) (*vsantypes.SuspendVApp_TaskResponse, error) {
  var reqBody, resBody SuspendVApp_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SuspendVM_TaskBody struct{
    Req *vsantypes.SuspendVM_Task `xml:"urn:vsan SuspendVM_Task,omitempty"`
    Res *vsantypes.SuspendVM_TaskResponse `xml:"urn:vsan SuspendVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SuspendVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func SuspendVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.SuspendVM_Task) (*vsantypes.SuspendVM_TaskResponse, error) {
  var reqBody, resBody SuspendVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type TerminateFaultTolerantVM_TaskBody struct{
    Req *vsantypes.TerminateFaultTolerantVM_Task `xml:"urn:vsan TerminateFaultTolerantVM_Task,omitempty"`
    Res *vsantypes.TerminateFaultTolerantVM_TaskResponse `xml:"urn:vsan TerminateFaultTolerantVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *TerminateFaultTolerantVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func TerminateFaultTolerantVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.TerminateFaultTolerantVM_Task) (*vsantypes.TerminateFaultTolerantVM_TaskResponse, error) {
  var reqBody, resBody TerminateFaultTolerantVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type TerminateProcessInGuestBody struct{
    Req *vsantypes.TerminateProcessInGuest `xml:"urn:vsan TerminateProcessInGuest,omitempty"`
    Res *vsantypes.TerminateProcessInGuestResponse `xml:"urn:vsan TerminateProcessInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *TerminateProcessInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func TerminateProcessInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.TerminateProcessInGuest) (*vsantypes.TerminateProcessInGuestResponse, error) {
  var reqBody, resBody TerminateProcessInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type TerminateSessionBody struct{
    Req *vsantypes.TerminateSession `xml:"urn:vsan TerminateSession,omitempty"`
    Res *vsantypes.TerminateSessionResponse `xml:"urn:vsan TerminateSessionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *TerminateSessionBody) Fault() *soap.Fault { return b.Fault_ }

func TerminateSession(ctx context.Context, r soap.RoundTripper, req *vsantypes.TerminateSession) (*vsantypes.TerminateSessionResponse, error) {
  var reqBody, resBody TerminateSessionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type TerminateVMBody struct{
    Req *vsantypes.TerminateVM `xml:"urn:vsan TerminateVM,omitempty"`
    Res *vsantypes.TerminateVMResponse `xml:"urn:vsan TerminateVMResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *TerminateVMBody) Fault() *soap.Fault { return b.Fault_ }

func TerminateVM(ctx context.Context, r soap.RoundTripper, req *vsantypes.TerminateVM) (*vsantypes.TerminateVMResponse, error) {
  var reqBody, resBody TerminateVMBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type TurnDiskLocatorLedOff_TaskBody struct{
    Req *vsantypes.TurnDiskLocatorLedOff_Task `xml:"urn:vsan TurnDiskLocatorLedOff_Task,omitempty"`
    Res *vsantypes.TurnDiskLocatorLedOff_TaskResponse `xml:"urn:vsan TurnDiskLocatorLedOff_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *TurnDiskLocatorLedOff_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func TurnDiskLocatorLedOff_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.TurnDiskLocatorLedOff_Task) (*vsantypes.TurnDiskLocatorLedOff_TaskResponse, error) {
  var reqBody, resBody TurnDiskLocatorLedOff_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type TurnDiskLocatorLedOn_TaskBody struct{
    Req *vsantypes.TurnDiskLocatorLedOn_Task `xml:"urn:vsan TurnDiskLocatorLedOn_Task,omitempty"`
    Res *vsantypes.TurnDiskLocatorLedOn_TaskResponse `xml:"urn:vsan TurnDiskLocatorLedOn_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *TurnDiskLocatorLedOn_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func TurnDiskLocatorLedOn_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.TurnDiskLocatorLedOn_Task) (*vsantypes.TurnDiskLocatorLedOn_TaskResponse, error) {
  var reqBody, resBody TurnDiskLocatorLedOn_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type TurnOffFaultToleranceForVM_TaskBody struct{
    Req *vsantypes.TurnOffFaultToleranceForVM_Task `xml:"urn:vsan TurnOffFaultToleranceForVM_Task,omitempty"`
    Res *vsantypes.TurnOffFaultToleranceForVM_TaskResponse `xml:"urn:vsan TurnOffFaultToleranceForVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *TurnOffFaultToleranceForVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func TurnOffFaultToleranceForVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.TurnOffFaultToleranceForVM_Task) (*vsantypes.TurnOffFaultToleranceForVM_TaskResponse, error) {
  var reqBody, resBody TurnOffFaultToleranceForVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnassignUserFromGroupBody struct{
    Req *vsantypes.UnassignUserFromGroup `xml:"urn:vsan UnassignUserFromGroup,omitempty"`
    Res *vsantypes.UnassignUserFromGroupResponse `xml:"urn:vsan UnassignUserFromGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnassignUserFromGroupBody) Fault() *soap.Fault { return b.Fault_ }

func UnassignUserFromGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnassignUserFromGroup) (*vsantypes.UnassignUserFromGroupResponse, error) {
  var reqBody, resBody UnassignUserFromGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnbindVnicBody struct{
    Req *vsantypes.UnbindVnic `xml:"urn:vsan UnbindVnic,omitempty"`
    Res *vsantypes.UnbindVnicResponse `xml:"urn:vsan UnbindVnicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnbindVnicBody) Fault() *soap.Fault { return b.Fault_ }

func UnbindVnic(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnbindVnic) (*vsantypes.UnbindVnicResponse, error) {
  var reqBody, resBody UnbindVnicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UninstallHostPatch_TaskBody struct{
    Req *vsantypes.UninstallHostPatch_Task `xml:"urn:vsan UninstallHostPatch_Task,omitempty"`
    Res *vsantypes.UninstallHostPatch_TaskResponse `xml:"urn:vsan UninstallHostPatch_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UninstallHostPatch_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UninstallHostPatch_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UninstallHostPatch_Task) (*vsantypes.UninstallHostPatch_TaskResponse, error) {
  var reqBody, resBody UninstallHostPatch_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UninstallIoFilter_TaskBody struct{
    Req *vsantypes.UninstallIoFilter_Task `xml:"urn:vsan UninstallIoFilter_Task,omitempty"`
    Res *vsantypes.UninstallIoFilter_TaskResponse `xml:"urn:vsan UninstallIoFilter_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UninstallIoFilter_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UninstallIoFilter_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UninstallIoFilter_Task) (*vsantypes.UninstallIoFilter_TaskResponse, error) {
  var reqBody, resBody UninstallIoFilter_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UninstallServiceBody struct{
    Req *vsantypes.UninstallService `xml:"urn:vsan UninstallService,omitempty"`
    Res *vsantypes.UninstallServiceResponse `xml:"urn:vsan UninstallServiceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UninstallServiceBody) Fault() *soap.Fault { return b.Fault_ }

func UninstallService(ctx context.Context, r soap.RoundTripper, req *vsantypes.UninstallService) (*vsantypes.UninstallServiceResponse, error) {
  var reqBody, resBody UninstallServiceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnmapVmfsVolumeEx_TaskBody struct{
    Req *vsantypes.UnmapVmfsVolumeEx_Task `xml:"urn:vsan UnmapVmfsVolumeEx_Task,omitempty"`
    Res *vsantypes.UnmapVmfsVolumeEx_TaskResponse `xml:"urn:vsan UnmapVmfsVolumeEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnmapVmfsVolumeEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UnmapVmfsVolumeEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnmapVmfsVolumeEx_Task) (*vsantypes.UnmapVmfsVolumeEx_TaskResponse, error) {
  var reqBody, resBody UnmapVmfsVolumeEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnmountDiskMapping_TaskBody struct{
    Req *vsantypes.UnmountDiskMapping_Task `xml:"urn:vsan UnmountDiskMapping_Task,omitempty"`
    Res *vsantypes.UnmountDiskMapping_TaskResponse `xml:"urn:vsan UnmountDiskMapping_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnmountDiskMapping_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UnmountDiskMapping_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnmountDiskMapping_Task) (*vsantypes.UnmountDiskMapping_TaskResponse, error) {
  var reqBody, resBody UnmountDiskMapping_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnmountForceMountedVmfsVolumeBody struct{
    Req *vsantypes.UnmountForceMountedVmfsVolume `xml:"urn:vsan UnmountForceMountedVmfsVolume,omitempty"`
    Res *vsantypes.UnmountForceMountedVmfsVolumeResponse `xml:"urn:vsan UnmountForceMountedVmfsVolumeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnmountForceMountedVmfsVolumeBody) Fault() *soap.Fault { return b.Fault_ }

func UnmountForceMountedVmfsVolume(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnmountForceMountedVmfsVolume) (*vsantypes.UnmountForceMountedVmfsVolumeResponse, error) {
  var reqBody, resBody UnmountForceMountedVmfsVolumeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnmountToolsInstallerBody struct{
    Req *vsantypes.UnmountToolsInstaller `xml:"urn:vsan UnmountToolsInstaller,omitempty"`
    Res *vsantypes.UnmountToolsInstallerResponse `xml:"urn:vsan UnmountToolsInstallerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnmountToolsInstallerBody) Fault() *soap.Fault { return b.Fault_ }

func UnmountToolsInstaller(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnmountToolsInstaller) (*vsantypes.UnmountToolsInstallerResponse, error) {
  var reqBody, resBody UnmountToolsInstallerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnmountVffsVolumeBody struct{
    Req *vsantypes.UnmountVffsVolume `xml:"urn:vsan UnmountVffsVolume,omitempty"`
    Res *vsantypes.UnmountVffsVolumeResponse `xml:"urn:vsan UnmountVffsVolumeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnmountVffsVolumeBody) Fault() *soap.Fault { return b.Fault_ }

func UnmountVffsVolume(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnmountVffsVolume) (*vsantypes.UnmountVffsVolumeResponse, error) {
  var reqBody, resBody UnmountVffsVolumeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnmountVmfsVolumeBody struct{
    Req *vsantypes.UnmountVmfsVolume `xml:"urn:vsan UnmountVmfsVolume,omitempty"`
    Res *vsantypes.UnmountVmfsVolumeResponse `xml:"urn:vsan UnmountVmfsVolumeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnmountVmfsVolumeBody) Fault() *soap.Fault { return b.Fault_ }

func UnmountVmfsVolume(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnmountVmfsVolume) (*vsantypes.UnmountVmfsVolumeResponse, error) {
  var reqBody, resBody UnmountVmfsVolumeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnmountVmfsVolumeEx_TaskBody struct{
    Req *vsantypes.UnmountVmfsVolumeEx_Task `xml:"urn:vsan UnmountVmfsVolumeEx_Task,omitempty"`
    Res *vsantypes.UnmountVmfsVolumeEx_TaskResponse `xml:"urn:vsan UnmountVmfsVolumeEx_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnmountVmfsVolumeEx_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UnmountVmfsVolumeEx_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnmountVmfsVolumeEx_Task) (*vsantypes.UnmountVmfsVolumeEx_TaskResponse, error) {
  var reqBody, resBody UnmountVmfsVolumeEx_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnregisterAndDestroy_TaskBody struct{
    Req *vsantypes.UnregisterAndDestroy_Task `xml:"urn:vsan UnregisterAndDestroy_Task,omitempty"`
    Res *vsantypes.UnregisterAndDestroy_TaskResponse `xml:"urn:vsan UnregisterAndDestroy_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnregisterAndDestroy_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UnregisterAndDestroy_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnregisterAndDestroy_Task) (*vsantypes.UnregisterAndDestroy_TaskResponse, error) {
  var reqBody, resBody UnregisterAndDestroy_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnregisterExtensionBody struct{
    Req *vsantypes.UnregisterExtension `xml:"urn:vsan UnregisterExtension,omitempty"`
    Res *vsantypes.UnregisterExtensionResponse `xml:"urn:vsan UnregisterExtensionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnregisterExtensionBody) Fault() *soap.Fault { return b.Fault_ }

func UnregisterExtension(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnregisterExtension) (*vsantypes.UnregisterExtensionResponse, error) {
  var reqBody, resBody UnregisterExtensionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnregisterHealthUpdateProviderBody struct{
    Req *vsantypes.UnregisterHealthUpdateProvider `xml:"urn:vsan UnregisterHealthUpdateProvider,omitempty"`
    Res *vsantypes.UnregisterHealthUpdateProviderResponse `xml:"urn:vsan UnregisterHealthUpdateProviderResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnregisterHealthUpdateProviderBody) Fault() *soap.Fault { return b.Fault_ }

func UnregisterHealthUpdateProvider(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnregisterHealthUpdateProvider) (*vsantypes.UnregisterHealthUpdateProviderResponse, error) {
  var reqBody, resBody UnregisterHealthUpdateProviderBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnregisterVMBody struct{
    Req *vsantypes.UnregisterVM `xml:"urn:vsan UnregisterVM,omitempty"`
    Res *vsantypes.UnregisterVMResponse `xml:"urn:vsan UnregisterVMResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnregisterVMBody) Fault() *soap.Fault { return b.Fault_ }

func UnregisterVM(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnregisterVM) (*vsantypes.UnregisterVMResponse, error) {
  var reqBody, resBody UnregisterVMBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateAnswerFile_TaskBody struct{
    Req *vsantypes.UpdateAnswerFile_Task `xml:"urn:vsan UpdateAnswerFile_Task,omitempty"`
    Res *vsantypes.UpdateAnswerFile_TaskResponse `xml:"urn:vsan UpdateAnswerFile_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateAnswerFile_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateAnswerFile_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateAnswerFile_Task) (*vsantypes.UpdateAnswerFile_TaskResponse, error) {
  var reqBody, resBody UpdateAnswerFile_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateAssignedLicenseBody struct{
    Req *vsantypes.UpdateAssignedLicense `xml:"urn:vsan UpdateAssignedLicense,omitempty"`
    Res *vsantypes.UpdateAssignedLicenseResponse `xml:"urn:vsan UpdateAssignedLicenseResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateAssignedLicenseBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateAssignedLicense(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateAssignedLicense) (*vsantypes.UpdateAssignedLicenseResponse, error) {
  var reqBody, resBody UpdateAssignedLicenseBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateAuthorizationRoleBody struct{
    Req *vsantypes.UpdateAuthorizationRole `xml:"urn:vsan UpdateAuthorizationRole,omitempty"`
    Res *vsantypes.UpdateAuthorizationRoleResponse `xml:"urn:vsan UpdateAuthorizationRoleResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateAuthorizationRoleBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateAuthorizationRole(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateAuthorizationRole) (*vsantypes.UpdateAuthorizationRoleResponse, error) {
  var reqBody, resBody UpdateAuthorizationRoleBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateBootDeviceBody struct{
    Req *vsantypes.UpdateBootDevice `xml:"urn:vsan UpdateBootDevice,omitempty"`
    Res *vsantypes.UpdateBootDeviceResponse `xml:"urn:vsan UpdateBootDeviceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateBootDeviceBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateBootDevice(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateBootDevice) (*vsantypes.UpdateBootDeviceResponse, error) {
  var reqBody, resBody UpdateBootDeviceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateChildResourceConfigurationBody struct{
    Req *vsantypes.UpdateChildResourceConfiguration `xml:"urn:vsan UpdateChildResourceConfiguration,omitempty"`
    Res *vsantypes.UpdateChildResourceConfigurationResponse `xml:"urn:vsan UpdateChildResourceConfigurationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateChildResourceConfigurationBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateChildResourceConfiguration(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateChildResourceConfiguration) (*vsantypes.UpdateChildResourceConfigurationResponse, error) {
  var reqBody, resBody UpdateChildResourceConfigurationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateClusterProfileBody struct{
    Req *vsantypes.UpdateClusterProfile `xml:"urn:vsan UpdateClusterProfile,omitempty"`
    Res *vsantypes.UpdateClusterProfileResponse `xml:"urn:vsan UpdateClusterProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateClusterProfileBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateClusterProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateClusterProfile) (*vsantypes.UpdateClusterProfileResponse, error) {
  var reqBody, resBody UpdateClusterProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateConfigBody struct{
    Req *vsantypes.UpdateConfig `xml:"urn:vsan UpdateConfig,omitempty"`
    Res *vsantypes.UpdateConfigResponse `xml:"urn:vsan UpdateConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateConfig) (*vsantypes.UpdateConfigResponse, error) {
  var reqBody, resBody UpdateConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateConsoleIpRouteConfigBody struct{
    Req *vsantypes.UpdateConsoleIpRouteConfig `xml:"urn:vsan UpdateConsoleIpRouteConfig,omitempty"`
    Res *vsantypes.UpdateConsoleIpRouteConfigResponse `xml:"urn:vsan UpdateConsoleIpRouteConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateConsoleIpRouteConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateConsoleIpRouteConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateConsoleIpRouteConfig) (*vsantypes.UpdateConsoleIpRouteConfigResponse, error) {
  var reqBody, resBody UpdateConsoleIpRouteConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateCounterLevelMappingBody struct{
    Req *vsantypes.UpdateCounterLevelMapping `xml:"urn:vsan UpdateCounterLevelMapping,omitempty"`
    Res *vsantypes.UpdateCounterLevelMappingResponse `xml:"urn:vsan UpdateCounterLevelMappingResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateCounterLevelMappingBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateCounterLevelMapping(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateCounterLevelMapping) (*vsantypes.UpdateCounterLevelMappingResponse, error) {
  var reqBody, resBody UpdateCounterLevelMappingBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDVSHealthCheckConfig_TaskBody struct{
    Req *vsantypes.UpdateDVSHealthCheckConfig_Task `xml:"urn:vsan UpdateDVSHealthCheckConfig_Task,omitempty"`
    Res *vsantypes.UpdateDVSHealthCheckConfig_TaskResponse `xml:"urn:vsan UpdateDVSHealthCheckConfig_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDVSHealthCheckConfig_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDVSHealthCheckConfig_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDVSHealthCheckConfig_Task) (*vsantypes.UpdateDVSHealthCheckConfig_TaskResponse, error) {
  var reqBody, resBody UpdateDVSHealthCheckConfig_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDVSLacpGroupConfig_TaskBody struct{
    Req *vsantypes.UpdateDVSLacpGroupConfig_Task `xml:"urn:vsan UpdateDVSLacpGroupConfig_Task,omitempty"`
    Res *vsantypes.UpdateDVSLacpGroupConfig_TaskResponse `xml:"urn:vsan UpdateDVSLacpGroupConfig_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDVSLacpGroupConfig_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDVSLacpGroupConfig_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDVSLacpGroupConfig_Task) (*vsantypes.UpdateDVSLacpGroupConfig_TaskResponse, error) {
  var reqBody, resBody UpdateDVSLacpGroupConfig_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDateTimeBody struct{
    Req *vsantypes.UpdateDateTime `xml:"urn:vsan UpdateDateTime,omitempty"`
    Res *vsantypes.UpdateDateTimeResponse `xml:"urn:vsan UpdateDateTimeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDateTimeBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDateTime(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDateTime) (*vsantypes.UpdateDateTimeResponse, error) {
  var reqBody, resBody UpdateDateTimeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDateTimeConfigBody struct{
    Req *vsantypes.UpdateDateTimeConfig `xml:"urn:vsan UpdateDateTimeConfig,omitempty"`
    Res *vsantypes.UpdateDateTimeConfigResponse `xml:"urn:vsan UpdateDateTimeConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDateTimeConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDateTimeConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDateTimeConfig) (*vsantypes.UpdateDateTimeConfigResponse, error) {
  var reqBody, resBody UpdateDateTimeConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDefaultPolicyBody struct{
    Req *vsantypes.UpdateDefaultPolicy `xml:"urn:vsan UpdateDefaultPolicy,omitempty"`
    Res *vsantypes.UpdateDefaultPolicyResponse `xml:"urn:vsan UpdateDefaultPolicyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDefaultPolicyBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDefaultPolicy(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDefaultPolicy) (*vsantypes.UpdateDefaultPolicyResponse, error) {
  var reqBody, resBody UpdateDefaultPolicyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDiskPartitionsBody struct{
    Req *vsantypes.UpdateDiskPartitions `xml:"urn:vsan UpdateDiskPartitions,omitempty"`
    Res *vsantypes.UpdateDiskPartitionsResponse `xml:"urn:vsan UpdateDiskPartitionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDiskPartitionsBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDiskPartitions(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDiskPartitions) (*vsantypes.UpdateDiskPartitionsResponse, error) {
  var reqBody, resBody UpdateDiskPartitionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDnsConfigBody struct{
    Req *vsantypes.UpdateDnsConfig `xml:"urn:vsan UpdateDnsConfig,omitempty"`
    Res *vsantypes.UpdateDnsConfigResponse `xml:"urn:vsan UpdateDnsConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDnsConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDnsConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDnsConfig) (*vsantypes.UpdateDnsConfigResponse, error) {
  var reqBody, resBody UpdateDnsConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateDvsCapabilityBody struct{
    Req *vsantypes.UpdateDvsCapability `xml:"urn:vsan UpdateDvsCapability,omitempty"`
    Res *vsantypes.UpdateDvsCapabilityResponse `xml:"urn:vsan UpdateDvsCapabilityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateDvsCapabilityBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateDvsCapability(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateDvsCapability) (*vsantypes.UpdateDvsCapabilityResponse, error) {
  var reqBody, resBody UpdateDvsCapabilityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateExtensionBody struct{
    Req *vsantypes.UpdateExtension `xml:"urn:vsan UpdateExtension,omitempty"`
    Res *vsantypes.UpdateExtensionResponse `xml:"urn:vsan UpdateExtensionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateExtensionBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateExtension(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateExtension) (*vsantypes.UpdateExtensionResponse, error) {
  var reqBody, resBody UpdateExtensionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateFlagsBody struct{
    Req *vsantypes.UpdateFlags `xml:"urn:vsan UpdateFlags,omitempty"`
    Res *vsantypes.UpdateFlagsResponse `xml:"urn:vsan UpdateFlagsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateFlagsBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateFlags(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateFlags) (*vsantypes.UpdateFlagsResponse, error) {
  var reqBody, resBody UpdateFlagsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateGraphicsConfigBody struct{
    Req *vsantypes.UpdateGraphicsConfig `xml:"urn:vsan UpdateGraphicsConfig,omitempty"`
    Res *vsantypes.UpdateGraphicsConfigResponse `xml:"urn:vsan UpdateGraphicsConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateGraphicsConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateGraphicsConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateGraphicsConfig) (*vsantypes.UpdateGraphicsConfigResponse, error) {
  var reqBody, resBody UpdateGraphicsConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateHostCustomizations_TaskBody struct{
    Req *vsantypes.UpdateHostCustomizations_Task `xml:"urn:vsan UpdateHostCustomizations_Task,omitempty"`
    Res *vsantypes.UpdateHostCustomizations_TaskResponse `xml:"urn:vsan UpdateHostCustomizations_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateHostCustomizations_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateHostCustomizations_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateHostCustomizations_Task) (*vsantypes.UpdateHostCustomizations_TaskResponse, error) {
  var reqBody, resBody UpdateHostCustomizations_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateHostImageAcceptanceLevelBody struct{
    Req *vsantypes.UpdateHostImageAcceptanceLevel `xml:"urn:vsan UpdateHostImageAcceptanceLevel,omitempty"`
    Res *vsantypes.UpdateHostImageAcceptanceLevelResponse `xml:"urn:vsan UpdateHostImageAcceptanceLevelResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateHostImageAcceptanceLevelBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateHostImageAcceptanceLevel(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateHostImageAcceptanceLevel) (*vsantypes.UpdateHostImageAcceptanceLevelResponse, error) {
  var reqBody, resBody UpdateHostImageAcceptanceLevelBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateHostProfileBody struct{
    Req *vsantypes.UpdateHostProfile `xml:"urn:vsan UpdateHostProfile,omitempty"`
    Res *vsantypes.UpdateHostProfileResponse `xml:"urn:vsan UpdateHostProfileResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateHostProfileBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateHostProfile(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateHostProfile) (*vsantypes.UpdateHostProfileResponse, error) {
  var reqBody, resBody UpdateHostProfileBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateHostSpecificationBody struct{
    Req *vsantypes.UpdateHostSpecification `xml:"urn:vsan UpdateHostSpecification,omitempty"`
    Res *vsantypes.UpdateHostSpecificationResponse `xml:"urn:vsan UpdateHostSpecificationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateHostSpecificationBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateHostSpecification(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateHostSpecification) (*vsantypes.UpdateHostSpecificationResponse, error) {
  var reqBody, resBody UpdateHostSpecificationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateHostSubSpecificationBody struct{
    Req *vsantypes.UpdateHostSubSpecification `xml:"urn:vsan UpdateHostSubSpecification,omitempty"`
    Res *vsantypes.UpdateHostSubSpecificationResponse `xml:"urn:vsan UpdateHostSubSpecificationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateHostSubSpecificationBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateHostSubSpecification(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateHostSubSpecification) (*vsantypes.UpdateHostSubSpecificationResponse, error) {
  var reqBody, resBody UpdateHostSubSpecificationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateInternetScsiAdvancedOptionsBody struct{
    Req *vsantypes.UpdateInternetScsiAdvancedOptions `xml:"urn:vsan UpdateInternetScsiAdvancedOptions,omitempty"`
    Res *vsantypes.UpdateInternetScsiAdvancedOptionsResponse `xml:"urn:vsan UpdateInternetScsiAdvancedOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateInternetScsiAdvancedOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateInternetScsiAdvancedOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateInternetScsiAdvancedOptions) (*vsantypes.UpdateInternetScsiAdvancedOptionsResponse, error) {
  var reqBody, resBody UpdateInternetScsiAdvancedOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateInternetScsiAliasBody struct{
    Req *vsantypes.UpdateInternetScsiAlias `xml:"urn:vsan UpdateInternetScsiAlias,omitempty"`
    Res *vsantypes.UpdateInternetScsiAliasResponse `xml:"urn:vsan UpdateInternetScsiAliasResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateInternetScsiAliasBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateInternetScsiAlias(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateInternetScsiAlias) (*vsantypes.UpdateInternetScsiAliasResponse, error) {
  var reqBody, resBody UpdateInternetScsiAliasBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateInternetScsiAuthenticationPropertiesBody struct{
    Req *vsantypes.UpdateInternetScsiAuthenticationProperties `xml:"urn:vsan UpdateInternetScsiAuthenticationProperties,omitempty"`
    Res *vsantypes.UpdateInternetScsiAuthenticationPropertiesResponse `xml:"urn:vsan UpdateInternetScsiAuthenticationPropertiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateInternetScsiAuthenticationPropertiesBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateInternetScsiAuthenticationProperties(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateInternetScsiAuthenticationProperties) (*vsantypes.UpdateInternetScsiAuthenticationPropertiesResponse, error) {
  var reqBody, resBody UpdateInternetScsiAuthenticationPropertiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateInternetScsiDigestPropertiesBody struct{
    Req *vsantypes.UpdateInternetScsiDigestProperties `xml:"urn:vsan UpdateInternetScsiDigestProperties,omitempty"`
    Res *vsantypes.UpdateInternetScsiDigestPropertiesResponse `xml:"urn:vsan UpdateInternetScsiDigestPropertiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateInternetScsiDigestPropertiesBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateInternetScsiDigestProperties(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateInternetScsiDigestProperties) (*vsantypes.UpdateInternetScsiDigestPropertiesResponse, error) {
  var reqBody, resBody UpdateInternetScsiDigestPropertiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateInternetScsiDiscoveryPropertiesBody struct{
    Req *vsantypes.UpdateInternetScsiDiscoveryProperties `xml:"urn:vsan UpdateInternetScsiDiscoveryProperties,omitempty"`
    Res *vsantypes.UpdateInternetScsiDiscoveryPropertiesResponse `xml:"urn:vsan UpdateInternetScsiDiscoveryPropertiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateInternetScsiDiscoveryPropertiesBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateInternetScsiDiscoveryProperties(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateInternetScsiDiscoveryProperties) (*vsantypes.UpdateInternetScsiDiscoveryPropertiesResponse, error) {
  var reqBody, resBody UpdateInternetScsiDiscoveryPropertiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateInternetScsiIPPropertiesBody struct{
    Req *vsantypes.UpdateInternetScsiIPProperties `xml:"urn:vsan UpdateInternetScsiIPProperties,omitempty"`
    Res *vsantypes.UpdateInternetScsiIPPropertiesResponse `xml:"urn:vsan UpdateInternetScsiIPPropertiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateInternetScsiIPPropertiesBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateInternetScsiIPProperties(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateInternetScsiIPProperties) (*vsantypes.UpdateInternetScsiIPPropertiesResponse, error) {
  var reqBody, resBody UpdateInternetScsiIPPropertiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateInternetScsiNameBody struct{
    Req *vsantypes.UpdateInternetScsiName `xml:"urn:vsan UpdateInternetScsiName,omitempty"`
    Res *vsantypes.UpdateInternetScsiNameResponse `xml:"urn:vsan UpdateInternetScsiNameResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateInternetScsiNameBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateInternetScsiName(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateInternetScsiName) (*vsantypes.UpdateInternetScsiNameResponse, error) {
  var reqBody, resBody UpdateInternetScsiNameBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateIpConfigBody struct{
    Req *vsantypes.UpdateIpConfig `xml:"urn:vsan UpdateIpConfig,omitempty"`
    Res *vsantypes.UpdateIpConfigResponse `xml:"urn:vsan UpdateIpConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateIpConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateIpConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateIpConfig) (*vsantypes.UpdateIpConfigResponse, error) {
  var reqBody, resBody UpdateIpConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateIpPoolBody struct{
    Req *vsantypes.UpdateIpPool `xml:"urn:vsan UpdateIpPool,omitempty"`
    Res *vsantypes.UpdateIpPoolResponse `xml:"urn:vsan UpdateIpPoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateIpPoolBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateIpPool(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateIpPool) (*vsantypes.UpdateIpPoolResponse, error) {
  var reqBody, resBody UpdateIpPoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateIpRouteConfigBody struct{
    Req *vsantypes.UpdateIpRouteConfig `xml:"urn:vsan UpdateIpRouteConfig,omitempty"`
    Res *vsantypes.UpdateIpRouteConfigResponse `xml:"urn:vsan UpdateIpRouteConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateIpRouteConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateIpRouteConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateIpRouteConfig) (*vsantypes.UpdateIpRouteConfigResponse, error) {
  var reqBody, resBody UpdateIpRouteConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateIpRouteTableConfigBody struct{
    Req *vsantypes.UpdateIpRouteTableConfig `xml:"urn:vsan UpdateIpRouteTableConfig,omitempty"`
    Res *vsantypes.UpdateIpRouteTableConfigResponse `xml:"urn:vsan UpdateIpRouteTableConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateIpRouteTableConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateIpRouteTableConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateIpRouteTableConfig) (*vsantypes.UpdateIpRouteTableConfigResponse, error) {
  var reqBody, resBody UpdateIpRouteTableConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateIpmiBody struct{
    Req *vsantypes.UpdateIpmi `xml:"urn:vsan UpdateIpmi,omitempty"`
    Res *vsantypes.UpdateIpmiResponse `xml:"urn:vsan UpdateIpmiResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateIpmiBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateIpmi(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateIpmi) (*vsantypes.UpdateIpmiResponse, error) {
  var reqBody, resBody UpdateIpmiBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateKmipServerBody struct{
    Req *vsantypes.UpdateKmipServer `xml:"urn:vsan UpdateKmipServer,omitempty"`
    Res *vsantypes.UpdateKmipServerResponse `xml:"urn:vsan UpdateKmipServerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateKmipServerBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateKmipServer(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateKmipServer) (*vsantypes.UpdateKmipServerResponse, error) {
  var reqBody, resBody UpdateKmipServerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateKmsSignedCsrClientCertBody struct{
    Req *vsantypes.UpdateKmsSignedCsrClientCert `xml:"urn:vsan UpdateKmsSignedCsrClientCert,omitempty"`
    Res *vsantypes.UpdateKmsSignedCsrClientCertResponse `xml:"urn:vsan UpdateKmsSignedCsrClientCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateKmsSignedCsrClientCertBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateKmsSignedCsrClientCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateKmsSignedCsrClientCert) (*vsantypes.UpdateKmsSignedCsrClientCertResponse, error) {
  var reqBody, resBody UpdateKmsSignedCsrClientCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateLicenseBody struct{
    Req *vsantypes.UpdateLicense `xml:"urn:vsan UpdateLicense,omitempty"`
    Res *vsantypes.UpdateLicenseResponse `xml:"urn:vsan UpdateLicenseResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateLicenseBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateLicense(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateLicense) (*vsantypes.UpdateLicenseResponse, error) {
  var reqBody, resBody UpdateLicenseBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateLicenseLabelBody struct{
    Req *vsantypes.UpdateLicenseLabel `xml:"urn:vsan UpdateLicenseLabel,omitempty"`
    Res *vsantypes.UpdateLicenseLabelResponse `xml:"urn:vsan UpdateLicenseLabelResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateLicenseLabelBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateLicenseLabel(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateLicenseLabel) (*vsantypes.UpdateLicenseLabelResponse, error) {
  var reqBody, resBody UpdateLicenseLabelBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateLinkedChildrenBody struct{
    Req *vsantypes.UpdateLinkedChildren `xml:"urn:vsan UpdateLinkedChildren,omitempty"`
    Res *vsantypes.UpdateLinkedChildrenResponse `xml:"urn:vsan UpdateLinkedChildrenResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateLinkedChildrenBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateLinkedChildren(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateLinkedChildren) (*vsantypes.UpdateLinkedChildrenResponse, error) {
  var reqBody, resBody UpdateLinkedChildrenBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateLocalSwapDatastoreBody struct{
    Req *vsantypes.UpdateLocalSwapDatastore `xml:"urn:vsan UpdateLocalSwapDatastore,omitempty"`
    Res *vsantypes.UpdateLocalSwapDatastoreResponse `xml:"urn:vsan UpdateLocalSwapDatastoreResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateLocalSwapDatastoreBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateLocalSwapDatastore(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateLocalSwapDatastore) (*vsantypes.UpdateLocalSwapDatastoreResponse, error) {
  var reqBody, resBody UpdateLocalSwapDatastoreBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateLockdownExceptionsBody struct{
    Req *vsantypes.UpdateLockdownExceptions `xml:"urn:vsan UpdateLockdownExceptions,omitempty"`
    Res *vsantypes.UpdateLockdownExceptionsResponse `xml:"urn:vsan UpdateLockdownExceptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateLockdownExceptionsBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateLockdownExceptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateLockdownExceptions) (*vsantypes.UpdateLockdownExceptionsResponse, error) {
  var reqBody, resBody UpdateLockdownExceptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateModuleOptionStringBody struct{
    Req *vsantypes.UpdateModuleOptionString `xml:"urn:vsan UpdateModuleOptionString,omitempty"`
    Res *vsantypes.UpdateModuleOptionStringResponse `xml:"urn:vsan UpdateModuleOptionStringResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateModuleOptionStringBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateModuleOptionString(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateModuleOptionString) (*vsantypes.UpdateModuleOptionStringResponse, error) {
  var reqBody, resBody UpdateModuleOptionStringBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateNetworkConfigBody struct{
    Req *vsantypes.UpdateNetworkConfig `xml:"urn:vsan UpdateNetworkConfig,omitempty"`
    Res *vsantypes.UpdateNetworkConfigResponse `xml:"urn:vsan UpdateNetworkConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateNetworkConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateNetworkConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateNetworkConfig) (*vsantypes.UpdateNetworkConfigResponse, error) {
  var reqBody, resBody UpdateNetworkConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateNetworkResourcePoolBody struct{
    Req *vsantypes.UpdateNetworkResourcePool `xml:"urn:vsan UpdateNetworkResourcePool,omitempty"`
    Res *vsantypes.UpdateNetworkResourcePoolResponse `xml:"urn:vsan UpdateNetworkResourcePoolResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateNetworkResourcePoolBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateNetworkResourcePool(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateNetworkResourcePool) (*vsantypes.UpdateNetworkResourcePoolResponse, error) {
  var reqBody, resBody UpdateNetworkResourcePoolBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateOptionsBody struct{
    Req *vsantypes.UpdateOptions `xml:"urn:vsan UpdateOptions,omitempty"`
    Res *vsantypes.UpdateOptionsResponse `xml:"urn:vsan UpdateOptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateOptionsBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateOptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateOptions) (*vsantypes.UpdateOptionsResponse, error) {
  var reqBody, resBody UpdateOptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdatePassthruConfigBody struct{
    Req *vsantypes.UpdatePassthruConfig `xml:"urn:vsan UpdatePassthruConfig,omitempty"`
    Res *vsantypes.UpdatePassthruConfigResponse `xml:"urn:vsan UpdatePassthruConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdatePassthruConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdatePassthruConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdatePassthruConfig) (*vsantypes.UpdatePassthruConfigResponse, error) {
  var reqBody, resBody UpdatePassthruConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdatePerfIntervalBody struct{
    Req *vsantypes.UpdatePerfInterval `xml:"urn:vsan UpdatePerfInterval,omitempty"`
    Res *vsantypes.UpdatePerfIntervalResponse `xml:"urn:vsan UpdatePerfIntervalResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdatePerfIntervalBody) Fault() *soap.Fault { return b.Fault_ }

func UpdatePerfInterval(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdatePerfInterval) (*vsantypes.UpdatePerfIntervalResponse, error) {
  var reqBody, resBody UpdatePerfIntervalBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdatePhysicalNicLinkSpeedBody struct{
    Req *vsantypes.UpdatePhysicalNicLinkSpeed `xml:"urn:vsan UpdatePhysicalNicLinkSpeed,omitempty"`
    Res *vsantypes.UpdatePhysicalNicLinkSpeedResponse `xml:"urn:vsan UpdatePhysicalNicLinkSpeedResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdatePhysicalNicLinkSpeedBody) Fault() *soap.Fault { return b.Fault_ }

func UpdatePhysicalNicLinkSpeed(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdatePhysicalNicLinkSpeed) (*vsantypes.UpdatePhysicalNicLinkSpeedResponse, error) {
  var reqBody, resBody UpdatePhysicalNicLinkSpeedBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdatePortGroupBody struct{
    Req *vsantypes.UpdatePortGroup `xml:"urn:vsan UpdatePortGroup,omitempty"`
    Res *vsantypes.UpdatePortGroupResponse `xml:"urn:vsan UpdatePortGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdatePortGroupBody) Fault() *soap.Fault { return b.Fault_ }

func UpdatePortGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdatePortGroup) (*vsantypes.UpdatePortGroupResponse, error) {
  var reqBody, resBody UpdatePortGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateProgressBody struct{
    Req *vsantypes.UpdateProgress `xml:"urn:vsan UpdateProgress,omitempty"`
    Res *vsantypes.UpdateProgressResponse `xml:"urn:vsan UpdateProgressResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateProgressBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateProgress(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateProgress) (*vsantypes.UpdateProgressResponse, error) {
  var reqBody, resBody UpdateProgressBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateReferenceHostBody struct{
    Req *vsantypes.UpdateReferenceHost `xml:"urn:vsan UpdateReferenceHost,omitempty"`
    Res *vsantypes.UpdateReferenceHostResponse `xml:"urn:vsan UpdateReferenceHostResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateReferenceHostBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateReferenceHost(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateReferenceHost) (*vsantypes.UpdateReferenceHostResponse, error) {
  var reqBody, resBody UpdateReferenceHostBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateRulesetBody struct{
    Req *vsantypes.UpdateRuleset `xml:"urn:vsan UpdateRuleset,omitempty"`
    Res *vsantypes.UpdateRulesetResponse `xml:"urn:vsan UpdateRulesetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateRulesetBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateRuleset(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateRuleset) (*vsantypes.UpdateRulesetResponse, error) {
  var reqBody, resBody UpdateRulesetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateScsiLunDisplayNameBody struct{
    Req *vsantypes.UpdateScsiLunDisplayName `xml:"urn:vsan UpdateScsiLunDisplayName,omitempty"`
    Res *vsantypes.UpdateScsiLunDisplayNameResponse `xml:"urn:vsan UpdateScsiLunDisplayNameResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateScsiLunDisplayNameBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateScsiLunDisplayName(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateScsiLunDisplayName) (*vsantypes.UpdateScsiLunDisplayNameResponse, error) {
  var reqBody, resBody UpdateScsiLunDisplayNameBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateSelfSignedClientCertBody struct{
    Req *vsantypes.UpdateSelfSignedClientCert `xml:"urn:vsan UpdateSelfSignedClientCert,omitempty"`
    Res *vsantypes.UpdateSelfSignedClientCertResponse `xml:"urn:vsan UpdateSelfSignedClientCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateSelfSignedClientCertBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateSelfSignedClientCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateSelfSignedClientCert) (*vsantypes.UpdateSelfSignedClientCertResponse, error) {
  var reqBody, resBody UpdateSelfSignedClientCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateServiceConsoleVirtualNicBody struct{
    Req *vsantypes.UpdateServiceConsoleVirtualNic `xml:"urn:vsan UpdateServiceConsoleVirtualNic,omitempty"`
    Res *vsantypes.UpdateServiceConsoleVirtualNicResponse `xml:"urn:vsan UpdateServiceConsoleVirtualNicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateServiceConsoleVirtualNicBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateServiceConsoleVirtualNic(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateServiceConsoleVirtualNic) (*vsantypes.UpdateServiceConsoleVirtualNicResponse, error) {
  var reqBody, resBody UpdateServiceConsoleVirtualNicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateServiceMessageBody struct{
    Req *vsantypes.UpdateServiceMessage `xml:"urn:vsan UpdateServiceMessage,omitempty"`
    Res *vsantypes.UpdateServiceMessageResponse `xml:"urn:vsan UpdateServiceMessageResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateServiceMessageBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateServiceMessage(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateServiceMessage) (*vsantypes.UpdateServiceMessageResponse, error) {
  var reqBody, resBody UpdateServiceMessageBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateServicePolicyBody struct{
    Req *vsantypes.UpdateServicePolicy `xml:"urn:vsan UpdateServicePolicy,omitempty"`
    Res *vsantypes.UpdateServicePolicyResponse `xml:"urn:vsan UpdateServicePolicyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateServicePolicyBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateServicePolicy(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateServicePolicy) (*vsantypes.UpdateServicePolicyResponse, error) {
  var reqBody, resBody UpdateServicePolicyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateSoftwareInternetScsiEnabledBody struct{
    Req *vsantypes.UpdateSoftwareInternetScsiEnabled `xml:"urn:vsan UpdateSoftwareInternetScsiEnabled,omitempty"`
    Res *vsantypes.UpdateSoftwareInternetScsiEnabledResponse `xml:"urn:vsan UpdateSoftwareInternetScsiEnabledResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateSoftwareInternetScsiEnabledBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateSoftwareInternetScsiEnabled(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateSoftwareInternetScsiEnabled) (*vsantypes.UpdateSoftwareInternetScsiEnabledResponse, error) {
  var reqBody, resBody UpdateSoftwareInternetScsiEnabledBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateSystemResourcesBody struct{
    Req *vsantypes.UpdateSystemResources `xml:"urn:vsan UpdateSystemResources,omitempty"`
    Res *vsantypes.UpdateSystemResourcesResponse `xml:"urn:vsan UpdateSystemResourcesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateSystemResourcesBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateSystemResources(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateSystemResources) (*vsantypes.UpdateSystemResourcesResponse, error) {
  var reqBody, resBody UpdateSystemResourcesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateSystemSwapConfigurationBody struct{
    Req *vsantypes.UpdateSystemSwapConfiguration `xml:"urn:vsan UpdateSystemSwapConfiguration,omitempty"`
    Res *vsantypes.UpdateSystemSwapConfigurationResponse `xml:"urn:vsan UpdateSystemSwapConfigurationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateSystemSwapConfigurationBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateSystemSwapConfiguration(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateSystemSwapConfiguration) (*vsantypes.UpdateSystemSwapConfigurationResponse, error) {
  var reqBody, resBody UpdateSystemSwapConfigurationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateSystemUsersBody struct{
    Req *vsantypes.UpdateSystemUsers `xml:"urn:vsan UpdateSystemUsers,omitempty"`
    Res *vsantypes.UpdateSystemUsersResponse `xml:"urn:vsan UpdateSystemUsersResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateSystemUsersBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateSystemUsers(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateSystemUsers) (*vsantypes.UpdateSystemUsersResponse, error) {
  var reqBody, resBody UpdateSystemUsersBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateUserBody struct{
    Req *vsantypes.UpdateUser `xml:"urn:vsan UpdateUser,omitempty"`
    Res *vsantypes.UpdateUserResponse `xml:"urn:vsan UpdateUserResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateUserBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateUser(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateUser) (*vsantypes.UpdateUserResponse, error) {
  var reqBody, resBody UpdateUserBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateVAppConfigBody struct{
    Req *vsantypes.UpdateVAppConfig `xml:"urn:vsan UpdateVAppConfig,omitempty"`
    Res *vsantypes.UpdateVAppConfigResponse `xml:"urn:vsan UpdateVAppConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateVAppConfigBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateVAppConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateVAppConfig) (*vsantypes.UpdateVAppConfigResponse, error) {
  var reqBody, resBody UpdateVAppConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateVVolVirtualMachineFiles_TaskBody struct{
    Req *vsantypes.UpdateVVolVirtualMachineFiles_Task `xml:"urn:vsan UpdateVVolVirtualMachineFiles_Task,omitempty"`
    Res *vsantypes.UpdateVVolVirtualMachineFiles_TaskResponse `xml:"urn:vsan UpdateVVolVirtualMachineFiles_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateVVolVirtualMachineFiles_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateVVolVirtualMachineFiles_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateVVolVirtualMachineFiles_Task) (*vsantypes.UpdateVVolVirtualMachineFiles_TaskResponse, error) {
  var reqBody, resBody UpdateVVolVirtualMachineFiles_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateVirtualMachineFiles_TaskBody struct{
    Req *vsantypes.UpdateVirtualMachineFiles_Task `xml:"urn:vsan UpdateVirtualMachineFiles_Task,omitempty"`
    Res *vsantypes.UpdateVirtualMachineFiles_TaskResponse `xml:"urn:vsan UpdateVirtualMachineFiles_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateVirtualMachineFiles_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateVirtualMachineFiles_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateVirtualMachineFiles_Task) (*vsantypes.UpdateVirtualMachineFiles_TaskResponse, error) {
  var reqBody, resBody UpdateVirtualMachineFiles_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateVirtualNicBody struct{
    Req *vsantypes.UpdateVirtualNic `xml:"urn:vsan UpdateVirtualNic,omitempty"`
    Res *vsantypes.UpdateVirtualNicResponse `xml:"urn:vsan UpdateVirtualNicResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateVirtualNicBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateVirtualNic(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateVirtualNic) (*vsantypes.UpdateVirtualNicResponse, error) {
  var reqBody, resBody UpdateVirtualNicBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateVirtualSwitchBody struct{
    Req *vsantypes.UpdateVirtualSwitch `xml:"urn:vsan UpdateVirtualSwitch,omitempty"`
    Res *vsantypes.UpdateVirtualSwitchResponse `xml:"urn:vsan UpdateVirtualSwitchResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateVirtualSwitchBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateVirtualSwitch(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateVirtualSwitch) (*vsantypes.UpdateVirtualSwitchResponse, error) {
  var reqBody, resBody UpdateVirtualSwitchBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateVmfsUnmapPriorityBody struct{
    Req *vsantypes.UpdateVmfsUnmapPriority `xml:"urn:vsan UpdateVmfsUnmapPriority,omitempty"`
    Res *vsantypes.UpdateVmfsUnmapPriorityResponse `xml:"urn:vsan UpdateVmfsUnmapPriorityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateVmfsUnmapPriorityBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateVmfsUnmapPriority(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateVmfsUnmapPriority) (*vsantypes.UpdateVmfsUnmapPriorityResponse, error) {
  var reqBody, resBody UpdateVmfsUnmapPriorityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpdateVsan_TaskBody struct{
    Req *vsantypes.UpdateVsan_Task `xml:"urn:vsan UpdateVsan_Task,omitempty"`
    Res *vsantypes.UpdateVsan_TaskResponse `xml:"urn:vsan UpdateVsan_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpdateVsan_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpdateVsan_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpdateVsan_Task) (*vsantypes.UpdateVsan_TaskResponse, error) {
  var reqBody, resBody UpdateVsan_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpgradeIoFilter_TaskBody struct{
    Req *vsantypes.UpgradeIoFilter_Task `xml:"urn:vsan UpgradeIoFilter_Task,omitempty"`
    Res *vsantypes.UpgradeIoFilter_TaskResponse `xml:"urn:vsan UpgradeIoFilter_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpgradeIoFilter_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpgradeIoFilter_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpgradeIoFilter_Task) (*vsantypes.UpgradeIoFilter_TaskResponse, error) {
  var reqBody, resBody UpgradeIoFilter_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpgradeTools_TaskBody struct{
    Req *vsantypes.UpgradeTools_Task `xml:"urn:vsan UpgradeTools_Task,omitempty"`
    Res *vsantypes.UpgradeTools_TaskResponse `xml:"urn:vsan UpgradeTools_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpgradeTools_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpgradeTools_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpgradeTools_Task) (*vsantypes.UpgradeTools_TaskResponse, error) {
  var reqBody, resBody UpgradeTools_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpgradeVM_TaskBody struct{
    Req *vsantypes.UpgradeVM_Task `xml:"urn:vsan UpgradeVM_Task,omitempty"`
    Res *vsantypes.UpgradeVM_TaskResponse `xml:"urn:vsan UpgradeVM_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpgradeVM_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UpgradeVM_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpgradeVM_Task) (*vsantypes.UpgradeVM_TaskResponse, error) {
  var reqBody, resBody UpgradeVM_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpgradeVmLayoutBody struct{
    Req *vsantypes.UpgradeVmLayout `xml:"urn:vsan UpgradeVmLayout,omitempty"`
    Res *vsantypes.UpgradeVmLayoutResponse `xml:"urn:vsan UpgradeVmLayoutResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpgradeVmLayoutBody) Fault() *soap.Fault { return b.Fault_ }

func UpgradeVmLayout(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpgradeVmLayout) (*vsantypes.UpgradeVmLayoutResponse, error) {
  var reqBody, resBody UpgradeVmLayoutBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpgradeVmfsBody struct{
    Req *vsantypes.UpgradeVmfs `xml:"urn:vsan UpgradeVmfs,omitempty"`
    Res *vsantypes.UpgradeVmfsResponse `xml:"urn:vsan UpgradeVmfsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpgradeVmfsBody) Fault() *soap.Fault { return b.Fault_ }

func UpgradeVmfs(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpgradeVmfs) (*vsantypes.UpgradeVmfsResponse, error) {
  var reqBody, resBody UpgradeVmfsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UpgradeVsanObjectsBody struct{
    Req *vsantypes.UpgradeVsanObjects `xml:"urn:vsan UpgradeVsanObjects,omitempty"`
    Res *vsantypes.UpgradeVsanObjectsResponse `xml:"urn:vsan UpgradeVsanObjectsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UpgradeVsanObjectsBody) Fault() *soap.Fault { return b.Fault_ }

func UpgradeVsanObjects(ctx context.Context, r soap.RoundTripper, req *vsantypes.UpgradeVsanObjects) (*vsantypes.UpgradeVsanObjectsResponse, error) {
  var reqBody, resBody UpgradeVsanObjectsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UploadClientCertBody struct{
    Req *vsantypes.UploadClientCert `xml:"urn:vsan UploadClientCert,omitempty"`
    Res *vsantypes.UploadClientCertResponse `xml:"urn:vsan UploadClientCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UploadClientCertBody) Fault() *soap.Fault { return b.Fault_ }

func UploadClientCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.UploadClientCert) (*vsantypes.UploadClientCertResponse, error) {
  var reqBody, resBody UploadClientCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UploadKmipServerCertBody struct{
    Req *vsantypes.UploadKmipServerCert `xml:"urn:vsan UploadKmipServerCert,omitempty"`
    Res *vsantypes.UploadKmipServerCertResponse `xml:"urn:vsan UploadKmipServerCertResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UploadKmipServerCertBody) Fault() *soap.Fault { return b.Fault_ }

func UploadKmipServerCert(ctx context.Context, r soap.RoundTripper, req *vsantypes.UploadKmipServerCert) (*vsantypes.UploadKmipServerCertResponse, error) {
  var reqBody, resBody UploadKmipServerCertBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANIsWitnessVirtualApplianceBody struct{
    Req *vsantypes.VSANIsWitnessVirtualAppliance `xml:"urn:vsan VSANIsWitnessVirtualAppliance,omitempty"`
    Res *vsantypes.VSANIsWitnessVirtualApplianceResponse `xml:"urn:vsan VSANIsWitnessVirtualApplianceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANIsWitnessVirtualApplianceBody) Fault() *soap.Fault { return b.Fault_ }

func VSANIsWitnessVirtualAppliance(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANIsWitnessVirtualAppliance) (*vsantypes.VSANIsWitnessVirtualApplianceResponse, error) {
  var reqBody, resBody VSANIsWitnessVirtualApplianceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANVcConvertToStretchedClusterBody struct{
    Req *vsantypes.VSANVcConvertToStretchedCluster `xml:"urn:vsan VSANVcConvertToStretchedCluster,omitempty"`
    Res *vsantypes.VSANVcConvertToStretchedClusterResponse `xml:"urn:vsan VSANVcConvertToStretchedClusterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANVcConvertToStretchedClusterBody) Fault() *soap.Fault { return b.Fault_ }

func VSANVcConvertToStretchedCluster(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANVcConvertToStretchedCluster) (*vsantypes.VSANVcConvertToStretchedClusterResponse, error) {
  var reqBody, resBody VSANVcConvertToStretchedClusterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANVcGetPreferredFaultDomainBody struct{
    Req *vsantypes.VSANVcGetPreferredFaultDomain `xml:"urn:vsan VSANVcGetPreferredFaultDomain,omitempty"`
    Res *vsantypes.VSANVcGetPreferredFaultDomainResponse `xml:"urn:vsan VSANVcGetPreferredFaultDomainResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANVcGetPreferredFaultDomainBody) Fault() *soap.Fault { return b.Fault_ }

func VSANVcGetPreferredFaultDomain(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANVcGetPreferredFaultDomain) (*vsantypes.VSANVcGetPreferredFaultDomainResponse, error) {
  var reqBody, resBody VSANVcGetPreferredFaultDomainBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANVcGetWitnessHostsBody struct{
    Req *vsantypes.VSANVcGetWitnessHosts `xml:"urn:vsan VSANVcGetWitnessHosts,omitempty"`
    Res *vsantypes.VSANVcGetWitnessHostsResponse `xml:"urn:vsan VSANVcGetWitnessHostsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANVcGetWitnessHostsBody) Fault() *soap.Fault { return b.Fault_ }

func VSANVcGetWitnessHosts(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANVcGetWitnessHosts) (*vsantypes.VSANVcGetWitnessHostsResponse, error) {
  var reqBody, resBody VSANVcGetWitnessHostsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANVcIsWitnessHostBody struct{
    Req *vsantypes.VSANVcIsWitnessHost `xml:"urn:vsan VSANVcIsWitnessHost,omitempty"`
    Res *vsantypes.VSANVcIsWitnessHostResponse `xml:"urn:vsan VSANVcIsWitnessHostResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANVcIsWitnessHostBody) Fault() *soap.Fault { return b.Fault_ }

func VSANVcIsWitnessHost(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANVcIsWitnessHost) (*vsantypes.VSANVcIsWitnessHostResponse, error) {
  var reqBody, resBody VSANVcIsWitnessHostBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANVcRemoveWitnessHostBody struct{
    Req *vsantypes.VSANVcRemoveWitnessHost `xml:"urn:vsan VSANVcRemoveWitnessHost,omitempty"`
    Res *vsantypes.VSANVcRemoveWitnessHostResponse `xml:"urn:vsan VSANVcRemoveWitnessHostResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANVcRemoveWitnessHostBody) Fault() *soap.Fault { return b.Fault_ }

func VSANVcRemoveWitnessHost(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANVcRemoveWitnessHost) (*vsantypes.VSANVcRemoveWitnessHostResponse, error) {
  var reqBody, resBody VSANVcRemoveWitnessHostBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANVcRetrieveStretchedClusterVcCapabilityBody struct{
    Req *vsantypes.VSANVcRetrieveStretchedClusterVcCapability `xml:"urn:vsan VSANVcRetrieveStretchedClusterVcCapability,omitempty"`
    Res *vsantypes.VSANVcRetrieveStretchedClusterVcCapabilityResponse `xml:"urn:vsan VSANVcRetrieveStretchedClusterVcCapabilityResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANVcRetrieveStretchedClusterVcCapabilityBody) Fault() *soap.Fault { return b.Fault_ }

func VSANVcRetrieveStretchedClusterVcCapability(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANVcRetrieveStretchedClusterVcCapability) (*vsantypes.VSANVcRetrieveStretchedClusterVcCapabilityResponse, error) {
  var reqBody, resBody VSANVcRetrieveStretchedClusterVcCapabilityBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VSANVcSetPreferredFaultDomainBody struct{
    Req *vsantypes.VSANVcSetPreferredFaultDomain `xml:"urn:vsan VSANVcSetPreferredFaultDomain,omitempty"`
    Res *vsantypes.VSANVcSetPreferredFaultDomainResponse `xml:"urn:vsan VSANVcSetPreferredFaultDomainResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VSANVcSetPreferredFaultDomainBody) Fault() *soap.Fault { return b.Fault_ }

func VSANVcSetPreferredFaultDomain(ctx context.Context, r soap.RoundTripper, req *vsantypes.VSANVcSetPreferredFaultDomain) (*vsantypes.VSANVcSetPreferredFaultDomainResponse, error) {
  var reqBody, resBody VSANVcSetPreferredFaultDomainBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ValidateCredentialsInGuestBody struct{
    Req *vsantypes.ValidateCredentialsInGuest `xml:"urn:vsan ValidateCredentialsInGuest,omitempty"`
    Res *vsantypes.ValidateCredentialsInGuestResponse `xml:"urn:vsan ValidateCredentialsInGuestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ValidateCredentialsInGuestBody) Fault() *soap.Fault { return b.Fault_ }

func ValidateCredentialsInGuest(ctx context.Context, r soap.RoundTripper, req *vsantypes.ValidateCredentialsInGuest) (*vsantypes.ValidateCredentialsInGuestResponse, error) {
  var reqBody, resBody ValidateCredentialsInGuestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ValidateHostBody struct{
    Req *vsantypes.ValidateHost `xml:"urn:vsan ValidateHost,omitempty"`
    Res *vsantypes.ValidateHostResponse `xml:"urn:vsan ValidateHostResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ValidateHostBody) Fault() *soap.Fault { return b.Fault_ }

func ValidateHost(ctx context.Context, r soap.RoundTripper, req *vsantypes.ValidateHost) (*vsantypes.ValidateHostResponse, error) {
  var reqBody, resBody ValidateHostBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ValidateMigrationBody struct{
    Req *vsantypes.ValidateMigration `xml:"urn:vsan ValidateMigration,omitempty"`
    Res *vsantypes.ValidateMigrationResponse `xml:"urn:vsan ValidateMigrationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ValidateMigrationBody) Fault() *soap.Fault { return b.Fault_ }

func ValidateMigration(ctx context.Context, r soap.RoundTripper, req *vsantypes.ValidateMigration) (*vsantypes.ValidateMigrationResponse, error) {
  var reqBody, resBody ValidateMigrationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VosQueryVsanObjectInformationBody struct{
    Req *vsantypes.VosQueryVsanObjectInformation `xml:"urn:vsan VosQueryVsanObjectInformation,omitempty"`
    Res *vsantypes.VosQueryVsanObjectInformationResponse `xml:"urn:vsan VosQueryVsanObjectInformationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VosQueryVsanObjectInformationBody) Fault() *soap.Fault { return b.Fault_ }

func VosQueryVsanObjectInformation(ctx context.Context, r soap.RoundTripper, req *vsantypes.VosQueryVsanObjectInformation) (*vsantypes.VosQueryVsanObjectInformationResponse, error) {
  var reqBody, resBody VosQueryVsanObjectInformationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VosSetVsanObjectPolicyBody struct{
    Req *vsantypes.VosSetVsanObjectPolicy `xml:"urn:vsan VosSetVsanObjectPolicy,omitempty"`
    Res *vsantypes.VosSetVsanObjectPolicyResponse `xml:"urn:vsan VosSetVsanObjectPolicyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VosSetVsanObjectPolicyBody) Fault() *soap.Fault { return b.Fault_ }

func VosSetVsanObjectPolicy(ctx context.Context, r soap.RoundTripper, req *vsantypes.VosSetVsanObjectPolicy) (*vsantypes.VosSetVsanObjectPolicyResponse, error) {
  var reqBody, resBody VosSetVsanObjectPolicyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanAttachVsanSupportBundleToSrBody struct{
    Req *vsantypes.VsanAttachVsanSupportBundleToSr `xml:"urn:vsan VsanAttachVsanSupportBundleToSr,omitempty"`
    Res *vsantypes.VsanAttachVsanSupportBundleToSrResponse `xml:"urn:vsan VsanAttachVsanSupportBundleToSrResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanAttachVsanSupportBundleToSrBody) Fault() *soap.Fault { return b.Fault_ }

func VsanAttachVsanSupportBundleToSr(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanAttachVsanSupportBundleToSr) (*vsantypes.VsanAttachVsanSupportBundleToSrResponse, error) {
  var reqBody, resBody VsanAttachVsanSupportBundleToSrBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanCheckClusterClomdLivenessBody struct{
    Req *vsantypes.VsanCheckClusterClomdLiveness `xml:"urn:vsan VsanCheckClusterClomdLiveness,omitempty"`
    Res *vsantypes.VsanCheckClusterClomdLivenessResponse `xml:"urn:vsan VsanCheckClusterClomdLivenessResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanCheckClusterClomdLivenessBody) Fault() *soap.Fault { return b.Fault_ }

func VsanCheckClusterClomdLiveness(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanCheckClusterClomdLiveness) (*vsantypes.VsanCheckClusterClomdLivenessResponse, error) {
  var reqBody, resBody VsanCheckClusterClomdLivenessBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanClusterGetConfigBody struct{
    Req *vsantypes.VsanClusterGetConfig `xml:"urn:vsan VsanClusterGetConfig,omitempty"`
    Res *vsantypes.VsanClusterGetConfigResponse `xml:"urn:vsan VsanClusterGetConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanClusterGetConfigBody) Fault() *soap.Fault { return b.Fault_ }

func VsanClusterGetConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanClusterGetConfig) (*vsantypes.VsanClusterGetConfigResponse, error) {
  var reqBody, resBody VsanClusterGetConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanClusterGetHclInfoBody struct{
    Req *vsantypes.VsanClusterGetHclInfo `xml:"urn:vsan VsanClusterGetHclInfo,omitempty"`
    Res *vsantypes.VsanClusterGetHclInfoResponse `xml:"urn:vsan VsanClusterGetHclInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanClusterGetHclInfoBody) Fault() *soap.Fault { return b.Fault_ }

func VsanClusterGetHclInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanClusterGetHclInfo) (*vsantypes.VsanClusterGetHclInfoResponse, error) {
  var reqBody, resBody VsanClusterGetHclInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanClusterGetRuntimeStatsBody struct{
    Req *vsantypes.VsanClusterGetRuntimeStats `xml:"urn:vsan VsanClusterGetRuntimeStats,omitempty"`
    Res *vsantypes.VsanClusterGetRuntimeStatsResponse `xml:"urn:vsan VsanClusterGetRuntimeStatsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanClusterGetRuntimeStatsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanClusterGetRuntimeStats(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanClusterGetRuntimeStats) (*vsantypes.VsanClusterGetRuntimeStatsResponse, error) {
  var reqBody, resBody VsanClusterGetRuntimeStatsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanClusterReconfigBody struct{
    Req *vsantypes.VsanClusterReconfig `xml:"urn:vsan VsanClusterReconfig,omitempty"`
    Res *vsantypes.VsanClusterReconfigResponse `xml:"urn:vsan VsanClusterReconfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanClusterReconfigBody) Fault() *soap.Fault { return b.Fault_ }

func VsanClusterReconfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanClusterReconfig) (*vsantypes.VsanClusterReconfigResponse, error) {
  var reqBody, resBody VsanClusterReconfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanCompleteMigrateVmsToVdsBody struct{
    Req *vsantypes.VsanCompleteMigrateVmsToVds `xml:"urn:vsan VsanCompleteMigrateVmsToVds,omitempty"`
    Res *vsantypes.VsanCompleteMigrateVmsToVdsResponse `xml:"urn:vsan VsanCompleteMigrateVmsToVdsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanCompleteMigrateVmsToVdsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanCompleteMigrateVmsToVds(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanCompleteMigrateVmsToVds) (*vsantypes.VsanCompleteMigrateVmsToVdsResponse, error) {
  var reqBody, resBody VsanCompleteMigrateVmsToVdsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanDownloadHclFile_TaskBody struct{
    Req *vsantypes.VsanDownloadHclFile_Task `xml:"urn:vsan VsanDownloadHclFile_Task,omitempty"`
    Res *vsantypes.VsanDownloadHclFile_TaskResponse `xml:"urn:vsan VsanDownloadHclFile_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanDownloadHclFile_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func VsanDownloadHclFile_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanDownloadHclFile_Task) (*vsantypes.VsanDownloadHclFile_TaskResponse, error) {
  var reqBody, resBody VsanDownloadHclFile_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanEncryptedClusterRekey_TaskBody struct{
    Req *vsantypes.VsanEncryptedClusterRekey_Task `xml:"urn:vsan VsanEncryptedClusterRekey_Task,omitempty"`
    Res *vsantypes.VsanEncryptedClusterRekey_TaskResponse `xml:"urn:vsan VsanEncryptedClusterRekey_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanEncryptedClusterRekey_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func VsanEncryptedClusterRekey_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanEncryptedClusterRekey_Task) (*vsantypes.VsanEncryptedClusterRekey_TaskResponse, error) {
  var reqBody, resBody VsanEncryptedClusterRekey_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanFlashScsiControllerFirmware_TaskBody struct{
    Req *vsantypes.VsanFlashScsiControllerFirmware_Task `xml:"urn:vsan VsanFlashScsiControllerFirmware_Task,omitempty"`
    Res *vsantypes.VsanFlashScsiControllerFirmware_TaskResponse `xml:"urn:vsan VsanFlashScsiControllerFirmware_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanFlashScsiControllerFirmware_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func VsanFlashScsiControllerFirmware_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanFlashScsiControllerFirmware_Task) (*vsantypes.VsanFlashScsiControllerFirmware_TaskResponse, error) {
  var reqBody, resBody VsanFlashScsiControllerFirmware_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanGetAboutInfoExBody struct{
    Req *vsantypes.VsanGetAboutInfoEx `xml:"urn:vsan VsanGetAboutInfoEx,omitempty"`
    Res *vsantypes.VsanGetAboutInfoExResponse `xml:"urn:vsan VsanGetAboutInfoExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanGetAboutInfoExBody) Fault() *soap.Fault { return b.Fault_ }

func VsanGetAboutInfoEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanGetAboutInfoEx) (*vsantypes.VsanGetAboutInfoExResponse, error) {
  var reqBody, resBody VsanGetAboutInfoExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanGetCapabilitiesBody struct{
    Req *vsantypes.VsanGetCapabilities `xml:"urn:vsan VsanGetCapabilities,omitempty"`
    Res *vsantypes.VsanGetCapabilitiesResponse `xml:"urn:vsan VsanGetCapabilitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanGetCapabilitiesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanGetCapabilities(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanGetCapabilities) (*vsantypes.VsanGetCapabilitiesResponse, error) {
  var reqBody, resBody VsanGetCapabilitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanGetHclInfoBody struct{
    Req *vsantypes.VsanGetHclInfo `xml:"urn:vsan VsanGetHclInfo,omitempty"`
    Res *vsantypes.VsanGetHclInfoResponse `xml:"urn:vsan VsanGetHclInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanGetHclInfoBody) Fault() *soap.Fault { return b.Fault_ }

func VsanGetHclInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanGetHclInfo) (*vsantypes.VsanGetHclInfoResponse, error) {
  var reqBody, resBody VsanGetHclInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanGetProactiveRebalanceInfoBody struct{
    Req *vsantypes.VsanGetProactiveRebalanceInfo `xml:"urn:vsan VsanGetProactiveRebalanceInfo,omitempty"`
    Res *vsantypes.VsanGetProactiveRebalanceInfoResponse `xml:"urn:vsan VsanGetProactiveRebalanceInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanGetProactiveRebalanceInfoBody) Fault() *soap.Fault { return b.Fault_ }

func VsanGetProactiveRebalanceInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanGetProactiveRebalanceInfo) (*vsantypes.VsanGetProactiveRebalanceInfoResponse, error) {
  var reqBody, resBody VsanGetProactiveRebalanceInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthGetVsanClusterSilentChecksBody struct{
    Req *vsantypes.VsanHealthGetVsanClusterSilentChecks `xml:"urn:vsan VsanHealthGetVsanClusterSilentChecks,omitempty"`
    Res *vsantypes.VsanHealthGetVsanClusterSilentChecksResponse `xml:"urn:vsan VsanHealthGetVsanClusterSilentChecksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthGetVsanClusterSilentChecksBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthGetVsanClusterSilentChecks(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthGetVsanClusterSilentChecks) (*vsantypes.VsanHealthGetVsanClusterSilentChecksResponse, error) {
  var reqBody, resBody VsanHealthGetVsanClusterSilentChecksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthIsRebalanceRunningBody struct{
    Req *vsantypes.VsanHealthIsRebalanceRunning `xml:"urn:vsan VsanHealthIsRebalanceRunning,omitempty"`
    Res *vsantypes.VsanHealthIsRebalanceRunningResponse `xml:"urn:vsan VsanHealthIsRebalanceRunningResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthIsRebalanceRunningBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthIsRebalanceRunning(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthIsRebalanceRunning) (*vsantypes.VsanHealthIsRebalanceRunningResponse, error) {
  var reqBody, resBody VsanHealthIsRebalanceRunningBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthQueryVsanClusterHealthCheckIntervalBody struct{
    Req *vsantypes.VsanHealthQueryVsanClusterHealthCheckInterval `xml:"urn:vsan VsanHealthQueryVsanClusterHealthCheckInterval,omitempty"`
    Res *vsantypes.VsanHealthQueryVsanClusterHealthCheckIntervalResponse `xml:"urn:vsan VsanHealthQueryVsanClusterHealthCheckIntervalResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthQueryVsanClusterHealthCheckIntervalBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthQueryVsanClusterHealthCheckInterval(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthQueryVsanClusterHealthCheckInterval) (*vsantypes.VsanHealthQueryVsanClusterHealthCheckIntervalResponse, error) {
  var reqBody, resBody VsanHealthQueryVsanClusterHealthCheckIntervalBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthQueryVsanClusterHealthConfigBody struct{
    Req *vsantypes.VsanHealthQueryVsanClusterHealthConfig `xml:"urn:vsan VsanHealthQueryVsanClusterHealthConfig,omitempty"`
    Res *vsantypes.VsanHealthQueryVsanClusterHealthConfigResponse `xml:"urn:vsan VsanHealthQueryVsanClusterHealthConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthQueryVsanClusterHealthConfigBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthQueryVsanClusterHealthConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthQueryVsanClusterHealthConfig) (*vsantypes.VsanHealthQueryVsanClusterHealthConfigResponse, error) {
  var reqBody, resBody VsanHealthQueryVsanClusterHealthConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthRepairClusterObjectsImmediateBody struct{
    Req *vsantypes.VsanHealthRepairClusterObjectsImmediate `xml:"urn:vsan VsanHealthRepairClusterObjectsImmediate,omitempty"`
    Res *vsantypes.VsanHealthRepairClusterObjectsImmediateResponse `xml:"urn:vsan VsanHealthRepairClusterObjectsImmediateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthRepairClusterObjectsImmediateBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthRepairClusterObjectsImmediate(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthRepairClusterObjectsImmediate) (*vsantypes.VsanHealthRepairClusterObjectsImmediateResponse, error) {
  var reqBody, resBody VsanHealthRepairClusterObjectsImmediateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthSendVsanTelemetryBody struct{
    Req *vsantypes.VsanHealthSendVsanTelemetry `xml:"urn:vsan VsanHealthSendVsanTelemetry,omitempty"`
    Res *vsantypes.VsanHealthSendVsanTelemetryResponse `xml:"urn:vsan VsanHealthSendVsanTelemetryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthSendVsanTelemetryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthSendVsanTelemetry(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthSendVsanTelemetry) (*vsantypes.VsanHealthSendVsanTelemetryResponse, error) {
  var reqBody, resBody VsanHealthSendVsanTelemetryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthSetLogLevelBody struct{
    Req *vsantypes.VsanHealthSetLogLevel `xml:"urn:vsan VsanHealthSetLogLevel,omitempty"`
    Res *vsantypes.VsanHealthSetLogLevelResponse `xml:"urn:vsan VsanHealthSetLogLevelResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthSetLogLevelBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthSetLogLevel(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthSetLogLevel) (*vsantypes.VsanHealthSetLogLevelResponse, error) {
  var reqBody, resBody VsanHealthSetLogLevelBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthSetVsanClusterHealthCheckIntervalBody struct{
    Req *vsantypes.VsanHealthSetVsanClusterHealthCheckInterval `xml:"urn:vsan VsanHealthSetVsanClusterHealthCheckInterval,omitempty"`
    Res *vsantypes.VsanHealthSetVsanClusterHealthCheckIntervalResponse `xml:"urn:vsan VsanHealthSetVsanClusterHealthCheckIntervalResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthSetVsanClusterHealthCheckIntervalBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthSetVsanClusterHealthCheckInterval(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthSetVsanClusterHealthCheckInterval) (*vsantypes.VsanHealthSetVsanClusterHealthCheckIntervalResponse, error) {
  var reqBody, resBody VsanHealthSetVsanClusterHealthCheckIntervalBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthSetVsanClusterSilentChecksBody struct{
    Req *vsantypes.VsanHealthSetVsanClusterSilentChecks `xml:"urn:vsan VsanHealthSetVsanClusterSilentChecks,omitempty"`
    Res *vsantypes.VsanHealthSetVsanClusterSilentChecksResponse `xml:"urn:vsan VsanHealthSetVsanClusterSilentChecksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthSetVsanClusterSilentChecksBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthSetVsanClusterSilentChecks(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthSetVsanClusterSilentChecks) (*vsantypes.VsanHealthSetVsanClusterSilentChecksResponse, error) {
  var reqBody, resBody VsanHealthSetVsanClusterSilentChecksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthSetVsanClusterTelemetryConfigBody struct{
    Req *vsantypes.VsanHealthSetVsanClusterTelemetryConfig `xml:"urn:vsan VsanHealthSetVsanClusterTelemetryConfig,omitempty"`
    Res *vsantypes.VsanHealthSetVsanClusterTelemetryConfigResponse `xml:"urn:vsan VsanHealthSetVsanClusterTelemetryConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthSetVsanClusterTelemetryConfigBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthSetVsanClusterTelemetryConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthSetVsanClusterTelemetryConfig) (*vsantypes.VsanHealthSetVsanClusterTelemetryConfigResponse, error) {
  var reqBody, resBody VsanHealthSetVsanClusterTelemetryConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHealthTestVsanClusterTelemetryProxyBody struct{
    Req *vsantypes.VsanHealthTestVsanClusterTelemetryProxy `xml:"urn:vsan VsanHealthTestVsanClusterTelemetryProxy,omitempty"`
    Res *vsantypes.VsanHealthTestVsanClusterTelemetryProxyResponse `xml:"urn:vsan VsanHealthTestVsanClusterTelemetryProxyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHealthTestVsanClusterTelemetryProxyBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHealthTestVsanClusterTelemetryProxy(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHealthTestVsanClusterTelemetryProxy) (*vsantypes.VsanHealthTestVsanClusterTelemetryProxyResponse, error) {
  var reqBody, resBody VsanHealthTestVsanClusterTelemetryProxyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostCleanupVmdkLoadTestBody struct{
    Req *vsantypes.VsanHostCleanupVmdkLoadTest `xml:"urn:vsan VsanHostCleanupVmdkLoadTest,omitempty"`
    Res *vsantypes.VsanHostCleanupVmdkLoadTestResponse `xml:"urn:vsan VsanHostCleanupVmdkLoadTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostCleanupVmdkLoadTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostCleanupVmdkLoadTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostCleanupVmdkLoadTest) (*vsantypes.VsanHostCleanupVmdkLoadTestResponse, error) {
  var reqBody, resBody VsanHostCleanupVmdkLoadTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostClomdLivenessBody struct{
    Req *vsantypes.VsanHostClomdLiveness `xml:"urn:vsan VsanHostClomdLiveness,omitempty"`
    Res *vsantypes.VsanHostClomdLivenessResponse `xml:"urn:vsan VsanHostClomdLivenessResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostClomdLivenessBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostClomdLiveness(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostClomdLiveness) (*vsantypes.VsanHostClomdLivenessResponse, error) {
  var reqBody, resBody VsanHostClomdLivenessBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostCreateVmHealthTestBody struct{
    Req *vsantypes.VsanHostCreateVmHealthTest `xml:"urn:vsan VsanHostCreateVmHealthTest,omitempty"`
    Res *vsantypes.VsanHostCreateVmHealthTestResponse `xml:"urn:vsan VsanHostCreateVmHealthTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostCreateVmHealthTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostCreateVmHealthTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostCreateVmHealthTest) (*vsantypes.VsanHostCreateVmHealthTestResponse, error) {
  var reqBody, resBody VsanHostCreateVmHealthTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostGetRuntimeStatsBody struct{
    Req *vsantypes.VsanHostGetRuntimeStats `xml:"urn:vsan VsanHostGetRuntimeStats,omitempty"`
    Res *vsantypes.VsanHostGetRuntimeStatsResponse `xml:"urn:vsan VsanHostGetRuntimeStatsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostGetRuntimeStatsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostGetRuntimeStats(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostGetRuntimeStats) (*vsantypes.VsanHostGetRuntimeStatsResponse, error) {
  var reqBody, resBody VsanHostGetRuntimeStatsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostPrepareVmdkLoadTestBody struct{
    Req *vsantypes.VsanHostPrepareVmdkLoadTest `xml:"urn:vsan VsanHostPrepareVmdkLoadTest,omitempty"`
    Res *vsantypes.VsanHostPrepareVmdkLoadTestResponse `xml:"urn:vsan VsanHostPrepareVmdkLoadTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostPrepareVmdkLoadTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostPrepareVmdkLoadTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostPrepareVmdkLoadTest) (*vsantypes.VsanHostPrepareVmdkLoadTestResponse, error) {
  var reqBody, resBody VsanHostPrepareVmdkLoadTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryAdvCfgBody struct{
    Req *vsantypes.VsanHostQueryAdvCfg `xml:"urn:vsan VsanHostQueryAdvCfg,omitempty"`
    Res *vsantypes.VsanHostQueryAdvCfgResponse `xml:"urn:vsan VsanHostQueryAdvCfgResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryAdvCfgBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryAdvCfg(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryAdvCfg) (*vsantypes.VsanHostQueryAdvCfgResponse, error) {
  var reqBody, resBody VsanHostQueryAdvCfgBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryCheckLimitsBody struct{
    Req *vsantypes.VsanHostQueryCheckLimits `xml:"urn:vsan VsanHostQueryCheckLimits,omitempty"`
    Res *vsantypes.VsanHostQueryCheckLimitsResponse `xml:"urn:vsan VsanHostQueryCheckLimitsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryCheckLimitsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryCheckLimits(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryCheckLimits) (*vsantypes.VsanHostQueryCheckLimitsResponse, error) {
  var reqBody, resBody VsanHostQueryCheckLimitsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryEncryptionHealthSummaryBody struct{
    Req *vsantypes.VsanHostQueryEncryptionHealthSummary `xml:"urn:vsan VsanHostQueryEncryptionHealthSummary,omitempty"`
    Res *vsantypes.VsanHostQueryEncryptionHealthSummaryResponse `xml:"urn:vsan VsanHostQueryEncryptionHealthSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryEncryptionHealthSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryEncryptionHealthSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryEncryptionHealthSummary) (*vsantypes.VsanHostQueryEncryptionHealthSummaryResponse, error) {
  var reqBody, resBody VsanHostQueryEncryptionHealthSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryHealthSystemVersionBody struct{
    Req *vsantypes.VsanHostQueryHealthSystemVersion `xml:"urn:vsan VsanHostQueryHealthSystemVersion,omitempty"`
    Res *vsantypes.VsanHostQueryHealthSystemVersionResponse `xml:"urn:vsan VsanHostQueryHealthSystemVersionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryHealthSystemVersionBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryHealthSystemVersion(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryHealthSystemVersion) (*vsantypes.VsanHostQueryHealthSystemVersionResponse, error) {
  var reqBody, resBody VsanHostQueryHealthSystemVersionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryHostInfoByUuidsBody struct{
    Req *vsantypes.VsanHostQueryHostInfoByUuids `xml:"urn:vsan VsanHostQueryHostInfoByUuids,omitempty"`
    Res *vsantypes.VsanHostQueryHostInfoByUuidsResponse `xml:"urn:vsan VsanHostQueryHostInfoByUuidsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryHostInfoByUuidsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryHostInfoByUuids(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryHostInfoByUuids) (*vsantypes.VsanHostQueryHostInfoByUuidsResponse, error) {
  var reqBody, resBody VsanHostQueryHostInfoByUuidsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryObjectHealthSummaryBody struct{
    Req *vsantypes.VsanHostQueryObjectHealthSummary `xml:"urn:vsan VsanHostQueryObjectHealthSummary,omitempty"`
    Res *vsantypes.VsanHostQueryObjectHealthSummaryResponse `xml:"urn:vsan VsanHostQueryObjectHealthSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryObjectHealthSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryObjectHealthSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryObjectHealthSummary) (*vsantypes.VsanHostQueryObjectHealthSummaryResponse, error) {
  var reqBody, resBody VsanHostQueryObjectHealthSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryPhysicalDiskHealthSummaryBody struct{
    Req *vsantypes.VsanHostQueryPhysicalDiskHealthSummary `xml:"urn:vsan VsanHostQueryPhysicalDiskHealthSummary,omitempty"`
    Res *vsantypes.VsanHostQueryPhysicalDiskHealthSummaryResponse `xml:"urn:vsan VsanHostQueryPhysicalDiskHealthSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryPhysicalDiskHealthSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryPhysicalDiskHealthSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryPhysicalDiskHealthSummary) (*vsantypes.VsanHostQueryPhysicalDiskHealthSummaryResponse, error) {
  var reqBody, resBody VsanHostQueryPhysicalDiskHealthSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryRunIperfClientBody struct{
    Req *vsantypes.VsanHostQueryRunIperfClient `xml:"urn:vsan VsanHostQueryRunIperfClient,omitempty"`
    Res *vsantypes.VsanHostQueryRunIperfClientResponse `xml:"urn:vsan VsanHostQueryRunIperfClientResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryRunIperfClientBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryRunIperfClient(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryRunIperfClient) (*vsantypes.VsanHostQueryRunIperfClientResponse, error) {
  var reqBody, resBody VsanHostQueryRunIperfClientBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryRunIperfServerBody struct{
    Req *vsantypes.VsanHostQueryRunIperfServer `xml:"urn:vsan VsanHostQueryRunIperfServer,omitempty"`
    Res *vsantypes.VsanHostQueryRunIperfServerResponse `xml:"urn:vsan VsanHostQueryRunIperfServerResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryRunIperfServerBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryRunIperfServer(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryRunIperfServer) (*vsantypes.VsanHostQueryRunIperfServerResponse, error) {
  var reqBody, resBody VsanHostQueryRunIperfServerBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQuerySmartStatsBody struct{
    Req *vsantypes.VsanHostQuerySmartStats `xml:"urn:vsan VsanHostQuerySmartStats,omitempty"`
    Res *vsantypes.VsanHostQuerySmartStatsResponse `xml:"urn:vsan VsanHostQuerySmartStatsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQuerySmartStatsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQuerySmartStats(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQuerySmartStats) (*vsantypes.VsanHostQuerySmartStatsResponse, error) {
  var reqBody, resBody VsanHostQuerySmartStatsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostQueryVerifyNetworkSettingsBody struct{
    Req *vsantypes.VsanHostQueryVerifyNetworkSettings `xml:"urn:vsan VsanHostQueryVerifyNetworkSettings,omitempty"`
    Res *vsantypes.VsanHostQueryVerifyNetworkSettingsResponse `xml:"urn:vsan VsanHostQueryVerifyNetworkSettingsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostQueryVerifyNetworkSettingsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostQueryVerifyNetworkSettings(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostQueryVerifyNetworkSettings) (*vsantypes.VsanHostQueryVerifyNetworkSettingsResponse, error) {
  var reqBody, resBody VsanHostQueryVerifyNetworkSettingsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostRepairImmediateObjectsBody struct{
    Req *vsantypes.VsanHostRepairImmediateObjects `xml:"urn:vsan VsanHostRepairImmediateObjects,omitempty"`
    Res *vsantypes.VsanHostRepairImmediateObjectsResponse `xml:"urn:vsan VsanHostRepairImmediateObjectsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostRepairImmediateObjectsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostRepairImmediateObjects(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostRepairImmediateObjects) (*vsantypes.VsanHostRepairImmediateObjectsResponse, error) {
  var reqBody, resBody VsanHostRepairImmediateObjectsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanHostRunVmdkLoadTestBody struct{
    Req *vsantypes.VsanHostRunVmdkLoadTest `xml:"urn:vsan VsanHostRunVmdkLoadTest,omitempty"`
    Res *vsantypes.VsanHostRunVmdkLoadTestResponse `xml:"urn:vsan VsanHostRunVmdkLoadTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanHostRunVmdkLoadTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanHostRunVmdkLoadTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanHostRunVmdkLoadTest) (*vsantypes.VsanHostRunVmdkLoadTestResponse, error) {
  var reqBody, resBody VsanHostRunVmdkLoadTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanMigrateVmsToVdsBody struct{
    Req *vsantypes.VsanMigrateVmsToVds `xml:"urn:vsan VsanMigrateVmsToVds,omitempty"`
    Res *vsantypes.VsanMigrateVmsToVdsResponse `xml:"urn:vsan VsanMigrateVmsToVdsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanMigrateVmsToVdsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanMigrateVmsToVds(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanMigrateVmsToVds) (*vsantypes.VsanMigrateVmsToVdsResponse, error) {
  var reqBody, resBody VsanMigrateVmsToVdsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfCreateStatsObjectBody struct{
    Req *vsantypes.VsanPerfCreateStatsObject `xml:"urn:vsan VsanPerfCreateStatsObject,omitempty"`
    Res *vsantypes.VsanPerfCreateStatsObjectResponse `xml:"urn:vsan VsanPerfCreateStatsObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfCreateStatsObjectBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfCreateStatsObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfCreateStatsObject) (*vsantypes.VsanPerfCreateStatsObjectResponse, error) {
  var reqBody, resBody VsanPerfCreateStatsObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfCreateStatsObjectTaskBody struct{
    Req *vsantypes.VsanPerfCreateStatsObjectTask `xml:"urn:vsan VsanPerfCreateStatsObjectTask,omitempty"`
    Res *vsantypes.VsanPerfCreateStatsObjectTaskResponse `xml:"urn:vsan VsanPerfCreateStatsObjectTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfCreateStatsObjectTaskBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfCreateStatsObjectTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfCreateStatsObjectTask) (*vsantypes.VsanPerfCreateStatsObjectTaskResponse, error) {
  var reqBody, resBody VsanPerfCreateStatsObjectTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfDeleteStatsObjectBody struct{
    Req *vsantypes.VsanPerfDeleteStatsObject `xml:"urn:vsan VsanPerfDeleteStatsObject,omitempty"`
    Res *vsantypes.VsanPerfDeleteStatsObjectResponse `xml:"urn:vsan VsanPerfDeleteStatsObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfDeleteStatsObjectBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfDeleteStatsObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfDeleteStatsObject) (*vsantypes.VsanPerfDeleteStatsObjectResponse, error) {
  var reqBody, resBody VsanPerfDeleteStatsObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfDeleteStatsObjectTaskBody struct{
    Req *vsantypes.VsanPerfDeleteStatsObjectTask `xml:"urn:vsan VsanPerfDeleteStatsObjectTask,omitempty"`
    Res *vsantypes.VsanPerfDeleteStatsObjectTaskResponse `xml:"urn:vsan VsanPerfDeleteStatsObjectTaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfDeleteStatsObjectTaskBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfDeleteStatsObjectTask(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfDeleteStatsObjectTask) (*vsantypes.VsanPerfDeleteStatsObjectTaskResponse, error) {
  var reqBody, resBody VsanPerfDeleteStatsObjectTaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfDeleteTimeRangeBody struct{
    Req *vsantypes.VsanPerfDeleteTimeRange `xml:"urn:vsan VsanPerfDeleteTimeRange,omitempty"`
    Res *vsantypes.VsanPerfDeleteTimeRangeResponse `xml:"urn:vsan VsanPerfDeleteTimeRangeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfDeleteTimeRangeBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfDeleteTimeRange(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfDeleteTimeRange) (*vsantypes.VsanPerfDeleteTimeRangeResponse, error) {
  var reqBody, resBody VsanPerfDeleteTimeRangeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfDiagnoseBody struct{
    Req *vsantypes.VsanPerfDiagnose `xml:"urn:vsan VsanPerfDiagnose,omitempty"`
    Res *vsantypes.VsanPerfDiagnoseResponse `xml:"urn:vsan VsanPerfDiagnoseResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfDiagnoseBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfDiagnose(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfDiagnose) (*vsantypes.VsanPerfDiagnoseResponse, error) {
  var reqBody, resBody VsanPerfDiagnoseBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfGetSupportedDiagnosticExceptionsBody struct{
    Req *vsantypes.VsanPerfGetSupportedDiagnosticExceptions `xml:"urn:vsan VsanPerfGetSupportedDiagnosticExceptions,omitempty"`
    Res *vsantypes.VsanPerfGetSupportedDiagnosticExceptionsResponse `xml:"urn:vsan VsanPerfGetSupportedDiagnosticExceptionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfGetSupportedDiagnosticExceptionsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfGetSupportedDiagnosticExceptions(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfGetSupportedDiagnosticExceptions) (*vsantypes.VsanPerfGetSupportedDiagnosticExceptionsResponse, error) {
  var reqBody, resBody VsanPerfGetSupportedDiagnosticExceptionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfGetSupportedEntityTypesBody struct{
    Req *vsantypes.VsanPerfGetSupportedEntityTypes `xml:"urn:vsan VsanPerfGetSupportedEntityTypes,omitempty"`
    Res *vsantypes.VsanPerfGetSupportedEntityTypesResponse `xml:"urn:vsan VsanPerfGetSupportedEntityTypesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfGetSupportedEntityTypesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfGetSupportedEntityTypes(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfGetSupportedEntityTypes) (*vsantypes.VsanPerfGetSupportedEntityTypesResponse, error) {
  var reqBody, resBody VsanPerfGetSupportedEntityTypesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfQueryClusterHealthBody struct{
    Req *vsantypes.VsanPerfQueryClusterHealth `xml:"urn:vsan VsanPerfQueryClusterHealth,omitempty"`
    Res *vsantypes.VsanPerfQueryClusterHealthResponse `xml:"urn:vsan VsanPerfQueryClusterHealthResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfQueryClusterHealthBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfQueryClusterHealth(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfQueryClusterHealth) (*vsantypes.VsanPerfQueryClusterHealthResponse, error) {
  var reqBody, resBody VsanPerfQueryClusterHealthBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfQueryNodeInformationBody struct{
    Req *vsantypes.VsanPerfQueryNodeInformation `xml:"urn:vsan VsanPerfQueryNodeInformation,omitempty"`
    Res *vsantypes.VsanPerfQueryNodeInformationResponse `xml:"urn:vsan VsanPerfQueryNodeInformationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfQueryNodeInformationBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfQueryNodeInformation(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfQueryNodeInformation) (*vsantypes.VsanPerfQueryNodeInformationResponse, error) {
  var reqBody, resBody VsanPerfQueryNodeInformationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfQueryPerfBody struct{
    Req *vsantypes.VsanPerfQueryPerf `xml:"urn:vsan VsanPerfQueryPerf,omitempty"`
    Res *vsantypes.VsanPerfQueryPerfResponse `xml:"urn:vsan VsanPerfQueryPerfResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfQueryPerfBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfQueryPerf(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfQueryPerf) (*vsantypes.VsanPerfQueryPerfResponse, error) {
  var reqBody, resBody VsanPerfQueryPerfBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfQueryStatsObjectInformationBody struct{
    Req *vsantypes.VsanPerfQueryStatsObjectInformation `xml:"urn:vsan VsanPerfQueryStatsObjectInformation,omitempty"`
    Res *vsantypes.VsanPerfQueryStatsObjectInformationResponse `xml:"urn:vsan VsanPerfQueryStatsObjectInformationResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfQueryStatsObjectInformationBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfQueryStatsObjectInformation(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfQueryStatsObjectInformation) (*vsantypes.VsanPerfQueryStatsObjectInformationResponse, error) {
  var reqBody, resBody VsanPerfQueryStatsObjectInformationBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfQueryTimeRangesBody struct{
    Req *vsantypes.VsanPerfQueryTimeRanges `xml:"urn:vsan VsanPerfQueryTimeRanges,omitempty"`
    Res *vsantypes.VsanPerfQueryTimeRangesResponse `xml:"urn:vsan VsanPerfQueryTimeRangesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfQueryTimeRangesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfQueryTimeRanges(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfQueryTimeRanges) (*vsantypes.VsanPerfQueryTimeRangesResponse, error) {
  var reqBody, resBody VsanPerfQueryTimeRangesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfSaveTimeRangesBody struct{
    Req *vsantypes.VsanPerfSaveTimeRanges `xml:"urn:vsan VsanPerfSaveTimeRanges,omitempty"`
    Res *vsantypes.VsanPerfSaveTimeRangesResponse `xml:"urn:vsan VsanPerfSaveTimeRangesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfSaveTimeRangesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfSaveTimeRanges(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfSaveTimeRanges) (*vsantypes.VsanPerfSaveTimeRangesResponse, error) {
  var reqBody, resBody VsanPerfSaveTimeRangesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfSetStatsObjectPolicyBody struct{
    Req *vsantypes.VsanPerfSetStatsObjectPolicy `xml:"urn:vsan VsanPerfSetStatsObjectPolicy,omitempty"`
    Res *vsantypes.VsanPerfSetStatsObjectPolicyResponse `xml:"urn:vsan VsanPerfSetStatsObjectPolicyResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfSetStatsObjectPolicyBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfSetStatsObjectPolicy(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfSetStatsObjectPolicy) (*vsantypes.VsanPerfSetStatsObjectPolicyResponse, error) {
  var reqBody, resBody VsanPerfSetStatsObjectPolicyBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerfToggleVerboseModeBody struct{
    Req *vsantypes.VsanPerfToggleVerboseMode `xml:"urn:vsan VsanPerfToggleVerboseMode,omitempty"`
    Res *vsantypes.VsanPerfToggleVerboseModeResponse `xml:"urn:vsan VsanPerfToggleVerboseModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerfToggleVerboseModeBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerfToggleVerboseMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerfToggleVerboseMode) (*vsantypes.VsanPerfToggleVerboseModeResponse, error) {
  var reqBody, resBody VsanPerfToggleVerboseModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPerformOnlineHealthCheckBody struct{
    Req *vsantypes.VsanPerformOnlineHealthCheck `xml:"urn:vsan VsanPerformOnlineHealthCheck,omitempty"`
    Res *vsantypes.VsanPerformOnlineHealthCheckResponse `xml:"urn:vsan VsanPerformOnlineHealthCheckResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPerformOnlineHealthCheckBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPerformOnlineHealthCheck(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPerformOnlineHealthCheck) (*vsantypes.VsanPerformOnlineHealthCheckResponse, error) {
  var reqBody, resBody VsanPerformOnlineHealthCheckBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPostConfigForVcsaBody struct{
    Req *vsantypes.VsanPostConfigForVcsa `xml:"urn:vsan VsanPostConfigForVcsa,omitempty"`
    Res *vsantypes.VsanPostConfigForVcsaResponse `xml:"urn:vsan VsanPostConfigForVcsaResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPostConfigForVcsaBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPostConfigForVcsa(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPostConfigForVcsa) (*vsantypes.VsanPostConfigForVcsaResponse, error) {
  var reqBody, resBody VsanPostConfigForVcsaBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPrepareVsanForVcsaBody struct{
    Req *vsantypes.VsanPrepareVsanForVcsa `xml:"urn:vsan VsanPrepareVsanForVcsa,omitempty"`
    Res *vsantypes.VsanPrepareVsanForVcsaResponse `xml:"urn:vsan VsanPrepareVsanForVcsaResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPrepareVsanForVcsaBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPrepareVsanForVcsa(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPrepareVsanForVcsa) (*vsantypes.VsanPrepareVsanForVcsaResponse, error) {
  var reqBody, resBody VsanPrepareVsanForVcsaBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanPurgeHclFilesBody struct{
    Req *vsantypes.VsanPurgeHclFiles `xml:"urn:vsan VsanPurgeHclFiles,omitempty"`
    Res *vsantypes.VsanPurgeHclFilesResponse `xml:"urn:vsan VsanPurgeHclFilesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanPurgeHclFilesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanPurgeHclFiles(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanPurgeHclFiles) (*vsantypes.VsanPurgeHclFilesResponse, error) {
  var reqBody, resBody VsanPurgeHclFilesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryAllSupportedHealthChecksBody struct{
    Req *vsantypes.VsanQueryAllSupportedHealthChecks `xml:"urn:vsan VsanQueryAllSupportedHealthChecks,omitempty"`
    Res *vsantypes.VsanQueryAllSupportedHealthChecksResponse `xml:"urn:vsan VsanQueryAllSupportedHealthChecksResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryAllSupportedHealthChecksBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryAllSupportedHealthChecks(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryAllSupportedHealthChecks) (*vsantypes.VsanQueryAllSupportedHealthChecksResponse, error) {
  var reqBody, resBody VsanQueryAllSupportedHealthChecksBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryAttachToSrHistoryBody struct{
    Req *vsantypes.VsanQueryAttachToSrHistory `xml:"urn:vsan VsanQueryAttachToSrHistory,omitempty"`
    Res *vsantypes.VsanQueryAttachToSrHistoryResponse `xml:"urn:vsan VsanQueryAttachToSrHistoryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryAttachToSrHistoryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryAttachToSrHistory(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryAttachToSrHistory) (*vsantypes.VsanQueryAttachToSrHistoryResponse, error) {
  var reqBody, resBody VsanQueryAttachToSrHistoryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryClusterAdvCfgSyncBody struct{
    Req *vsantypes.VsanQueryClusterAdvCfgSync `xml:"urn:vsan VsanQueryClusterAdvCfgSync,omitempty"`
    Res *vsantypes.VsanQueryClusterAdvCfgSyncResponse `xml:"urn:vsan VsanQueryClusterAdvCfgSyncResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryClusterAdvCfgSyncBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryClusterAdvCfgSync(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryClusterAdvCfgSync) (*vsantypes.VsanQueryClusterAdvCfgSyncResponse, error) {
  var reqBody, resBody VsanQueryClusterAdvCfgSyncBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryClusterCaptureVsanPcapBody struct{
    Req *vsantypes.VsanQueryClusterCaptureVsanPcap `xml:"urn:vsan VsanQueryClusterCaptureVsanPcap,omitempty"`
    Res *vsantypes.VsanQueryClusterCaptureVsanPcapResponse `xml:"urn:vsan VsanQueryClusterCaptureVsanPcapResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryClusterCaptureVsanPcapBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryClusterCaptureVsanPcap(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryClusterCaptureVsanPcap) (*vsantypes.VsanQueryClusterCaptureVsanPcapResponse, error) {
  var reqBody, resBody VsanQueryClusterCaptureVsanPcapBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryClusterCheckLimitsBody struct{
    Req *vsantypes.VsanQueryClusterCheckLimits `xml:"urn:vsan VsanQueryClusterCheckLimits,omitempty"`
    Res *vsantypes.VsanQueryClusterCheckLimitsResponse `xml:"urn:vsan VsanQueryClusterCheckLimitsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryClusterCheckLimitsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryClusterCheckLimits(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryClusterCheckLimits) (*vsantypes.VsanQueryClusterCheckLimitsResponse, error) {
  var reqBody, resBody VsanQueryClusterCheckLimitsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryClusterCreateVmHealthTestBody struct{
    Req *vsantypes.VsanQueryClusterCreateVmHealthTest `xml:"urn:vsan VsanQueryClusterCreateVmHealthTest,omitempty"`
    Res *vsantypes.VsanQueryClusterCreateVmHealthTestResponse `xml:"urn:vsan VsanQueryClusterCreateVmHealthTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryClusterCreateVmHealthTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryClusterCreateVmHealthTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryClusterCreateVmHealthTest) (*vsantypes.VsanQueryClusterCreateVmHealthTestResponse, error) {
  var reqBody, resBody VsanQueryClusterCreateVmHealthTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryClusterHealthSystemVersionsBody struct{
    Req *vsantypes.VsanQueryClusterHealthSystemVersions `xml:"urn:vsan VsanQueryClusterHealthSystemVersions,omitempty"`
    Res *vsantypes.VsanQueryClusterHealthSystemVersionsResponse `xml:"urn:vsan VsanQueryClusterHealthSystemVersionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryClusterHealthSystemVersionsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryClusterHealthSystemVersions(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryClusterHealthSystemVersions) (*vsantypes.VsanQueryClusterHealthSystemVersionsResponse, error) {
  var reqBody, resBody VsanQueryClusterHealthSystemVersionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryClusterNetworkPerfTestBody struct{
    Req *vsantypes.VsanQueryClusterNetworkPerfTest `xml:"urn:vsan VsanQueryClusterNetworkPerfTest,omitempty"`
    Res *vsantypes.VsanQueryClusterNetworkPerfTestResponse `xml:"urn:vsan VsanQueryClusterNetworkPerfTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryClusterNetworkPerfTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryClusterNetworkPerfTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryClusterNetworkPerfTest) (*vsantypes.VsanQueryClusterNetworkPerfTestResponse, error) {
  var reqBody, resBody VsanQueryClusterNetworkPerfTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryClusterPhysicalDiskHealthSummaryBody struct{
    Req *vsantypes.VsanQueryClusterPhysicalDiskHealthSummary `xml:"urn:vsan VsanQueryClusterPhysicalDiskHealthSummary,omitempty"`
    Res *vsantypes.VsanQueryClusterPhysicalDiskHealthSummaryResponse `xml:"urn:vsan VsanQueryClusterPhysicalDiskHealthSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryClusterPhysicalDiskHealthSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryClusterPhysicalDiskHealthSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryClusterPhysicalDiskHealthSummary) (*vsantypes.VsanQueryClusterPhysicalDiskHealthSummaryResponse, error) {
  var reqBody, resBody VsanQueryClusterPhysicalDiskHealthSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryObjectIdentitiesBody struct{
    Req *vsantypes.VsanQueryObjectIdentities `xml:"urn:vsan VsanQueryObjectIdentities,omitempty"`
    Res *vsantypes.VsanQueryObjectIdentitiesResponse `xml:"urn:vsan VsanQueryObjectIdentitiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryObjectIdentitiesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryObjectIdentities(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryObjectIdentities) (*vsantypes.VsanQueryObjectIdentitiesResponse, error) {
  var reqBody, resBody VsanQueryObjectIdentitiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQuerySpaceUsageBody struct{
    Req *vsantypes.VsanQuerySpaceUsage `xml:"urn:vsan VsanQuerySpaceUsage,omitempty"`
    Res *vsantypes.VsanQuerySpaceUsageResponse `xml:"urn:vsan VsanQuerySpaceUsageResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQuerySpaceUsageBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQuerySpaceUsage(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQuerySpaceUsage) (*vsantypes.VsanQuerySpaceUsageResponse, error) {
  var reqBody, resBody VsanQuerySpaceUsageBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryUpgradeStatusExBody struct{
    Req *vsantypes.VsanQueryUpgradeStatusEx `xml:"urn:vsan VsanQueryUpgradeStatusEx,omitempty"`
    Res *vsantypes.VsanQueryUpgradeStatusExResponse `xml:"urn:vsan VsanQueryUpgradeStatusExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryUpgradeStatusExBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryUpgradeStatusEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryUpgradeStatusEx) (*vsantypes.VsanQueryUpgradeStatusExResponse, error) {
  var reqBody, resBody VsanQueryUpgradeStatusExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterCreateVmHealthHistoryTestBody struct{
    Req *vsantypes.VsanQueryVcClusterCreateVmHealthHistoryTest `xml:"urn:vsan VsanQueryVcClusterCreateVmHealthHistoryTest,omitempty"`
    Res *vsantypes.VsanQueryVcClusterCreateVmHealthHistoryTestResponse `xml:"urn:vsan VsanQueryVcClusterCreateVmHealthHistoryTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterCreateVmHealthHistoryTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterCreateVmHealthHistoryTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterCreateVmHealthHistoryTest) (*vsantypes.VsanQueryVcClusterCreateVmHealthHistoryTestResponse, error) {
  var reqBody, resBody VsanQueryVcClusterCreateVmHealthHistoryTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterCreateVmHealthTestBody struct{
    Req *vsantypes.VsanQueryVcClusterCreateVmHealthTest `xml:"urn:vsan VsanQueryVcClusterCreateVmHealthTest,omitempty"`
    Res *vsantypes.VsanQueryVcClusterCreateVmHealthTestResponse `xml:"urn:vsan VsanQueryVcClusterCreateVmHealthTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterCreateVmHealthTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterCreateVmHealthTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterCreateVmHealthTest) (*vsantypes.VsanQueryVcClusterCreateVmHealthTestResponse, error) {
  var reqBody, resBody VsanQueryVcClusterCreateVmHealthTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterHealthSummaryBody struct{
    Req *vsantypes.VsanQueryVcClusterHealthSummary `xml:"urn:vsan VsanQueryVcClusterHealthSummary,omitempty"`
    Res *vsantypes.VsanQueryVcClusterHealthSummaryResponse `xml:"urn:vsan VsanQueryVcClusterHealthSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterHealthSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterHealthSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterHealthSummary) (*vsantypes.VsanQueryVcClusterHealthSummaryResponse, error) {
  var reqBody, resBody VsanQueryVcClusterHealthSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterNetworkPerfHistoryTestBody struct{
    Req *vsantypes.VsanQueryVcClusterNetworkPerfHistoryTest `xml:"urn:vsan VsanQueryVcClusterNetworkPerfHistoryTest,omitempty"`
    Res *vsantypes.VsanQueryVcClusterNetworkPerfHistoryTestResponse `xml:"urn:vsan VsanQueryVcClusterNetworkPerfHistoryTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterNetworkPerfHistoryTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterNetworkPerfHistoryTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterNetworkPerfHistoryTest) (*vsantypes.VsanQueryVcClusterNetworkPerfHistoryTestResponse, error) {
  var reqBody, resBody VsanQueryVcClusterNetworkPerfHistoryTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterNetworkPerfTestBody struct{
    Req *vsantypes.VsanQueryVcClusterNetworkPerfTest `xml:"urn:vsan VsanQueryVcClusterNetworkPerfTest,omitempty"`
    Res *vsantypes.VsanQueryVcClusterNetworkPerfTestResponse `xml:"urn:vsan VsanQueryVcClusterNetworkPerfTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterNetworkPerfTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterNetworkPerfTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterNetworkPerfTest) (*vsantypes.VsanQueryVcClusterNetworkPerfTestResponse, error) {
  var reqBody, resBody VsanQueryVcClusterNetworkPerfTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterSmartStatsSummaryBody struct{
    Req *vsantypes.VsanQueryVcClusterSmartStatsSummary `xml:"urn:vsan VsanQueryVcClusterSmartStatsSummary,omitempty"`
    Res *vsantypes.VsanQueryVcClusterSmartStatsSummaryResponse `xml:"urn:vsan VsanQueryVcClusterSmartStatsSummaryResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterSmartStatsSummaryBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterSmartStatsSummary(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterSmartStatsSummary) (*vsantypes.VsanQueryVcClusterSmartStatsSummaryResponse, error) {
  var reqBody, resBody VsanQueryVcClusterSmartStatsSummaryBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterVmdkLoadHistoryTestBody struct{
    Req *vsantypes.VsanQueryVcClusterVmdkLoadHistoryTest `xml:"urn:vsan VsanQueryVcClusterVmdkLoadHistoryTest,omitempty"`
    Res *vsantypes.VsanQueryVcClusterVmdkLoadHistoryTestResponse `xml:"urn:vsan VsanQueryVcClusterVmdkLoadHistoryTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterVmdkLoadHistoryTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterVmdkLoadHistoryTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterVmdkLoadHistoryTest) (*vsantypes.VsanQueryVcClusterVmdkLoadHistoryTestResponse, error) {
  var reqBody, resBody VsanQueryVcClusterVmdkLoadHistoryTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVcClusterVmdkWorkloadTypesBody struct{
    Req *vsantypes.VsanQueryVcClusterVmdkWorkloadTypes `xml:"urn:vsan VsanQueryVcClusterVmdkWorkloadTypes,omitempty"`
    Res *vsantypes.VsanQueryVcClusterVmdkWorkloadTypesResponse `xml:"urn:vsan VsanQueryVcClusterVmdkWorkloadTypesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVcClusterVmdkWorkloadTypesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVcClusterVmdkWorkloadTypes(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVcClusterVmdkWorkloadTypes) (*vsantypes.VsanQueryVcClusterVmdkWorkloadTypesResponse, error) {
  var reqBody, resBody VsanQueryVcClusterVmdkWorkloadTypesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryVerifyClusterNetworkSettingsBody struct{
    Req *vsantypes.VsanQueryVerifyClusterNetworkSettings `xml:"urn:vsan VsanQueryVerifyClusterNetworkSettings,omitempty"`
    Res *vsantypes.VsanQueryVerifyClusterNetworkSettingsResponse `xml:"urn:vsan VsanQueryVerifyClusterNetworkSettingsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryVerifyClusterNetworkSettingsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryVerifyClusterNetworkSettings(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryVerifyClusterNetworkSettings) (*vsantypes.VsanQueryVerifyClusterNetworkSettingsResponse, error) {
  var reqBody, resBody VsanQueryVerifyClusterNetworkSettingsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanQueryWhatIfEvacuationResultBody struct{
    Req *vsantypes.VsanQueryWhatIfEvacuationResult `xml:"urn:vsan VsanQueryWhatIfEvacuationResult,omitempty"`
    Res *vsantypes.VsanQueryWhatIfEvacuationResultResponse `xml:"urn:vsan VsanQueryWhatIfEvacuationResultResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanQueryWhatIfEvacuationResultBody) Fault() *soap.Fault { return b.Fault_ }

func VsanQueryWhatIfEvacuationResult(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanQueryWhatIfEvacuationResult) (*vsantypes.VsanQueryWhatIfEvacuationResultResponse, error) {
  var reqBody, resBody VsanQueryWhatIfEvacuationResultBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanRebalanceClusterBody struct{
    Req *vsantypes.VsanRebalanceCluster `xml:"urn:vsan VsanRebalanceCluster,omitempty"`
    Res *vsantypes.VsanRebalanceClusterResponse `xml:"urn:vsan VsanRebalanceClusterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanRebalanceClusterBody) Fault() *soap.Fault { return b.Fault_ }

func VsanRebalanceCluster(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanRebalanceCluster) (*vsantypes.VsanRebalanceClusterResponse, error) {
  var reqBody, resBody VsanRebalanceClusterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanRepairClusterImmediateObjectsBody struct{
    Req *vsantypes.VsanRepairClusterImmediateObjects `xml:"urn:vsan VsanRepairClusterImmediateObjects,omitempty"`
    Res *vsantypes.VsanRepairClusterImmediateObjectsResponse `xml:"urn:vsan VsanRepairClusterImmediateObjectsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanRepairClusterImmediateObjectsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanRepairClusterImmediateObjects(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanRepairClusterImmediateObjects) (*vsantypes.VsanRepairClusterImmediateObjectsResponse, error) {
  var reqBody, resBody VsanRepairClusterImmediateObjectsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanRetrievePropertiesBody struct{
    Req *vsantypes.VsanRetrieveProperties `xml:"urn:vsan VsanRetrieveProperties,omitempty"`
    Res *vsantypes.VsanRetrievePropertiesResponse `xml:"urn:vsan VsanRetrievePropertiesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanRetrievePropertiesBody) Fault() *soap.Fault { return b.Fault_ }

func VsanRetrieveProperties(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanRetrieveProperties) (*vsantypes.VsanRetrievePropertiesResponse, error) {
  var reqBody, resBody VsanRetrievePropertiesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanRollbackVdsToVssBody struct{
    Req *vsantypes.VsanRollbackVdsToVss `xml:"urn:vsan VsanRollbackVdsToVss,omitempty"`
    Res *vsantypes.VsanRollbackVdsToVssResponse `xml:"urn:vsan VsanRollbackVdsToVssResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanRollbackVdsToVssBody) Fault() *soap.Fault { return b.Fault_ }

func VsanRollbackVdsToVss(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanRollbackVdsToVss) (*vsantypes.VsanRollbackVdsToVssResponse, error) {
  var reqBody, resBody VsanRollbackVdsToVssBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanStartProactiveRebalanceBody struct{
    Req *vsantypes.VsanStartProactiveRebalance `xml:"urn:vsan VsanStartProactiveRebalance,omitempty"`
    Res *vsantypes.VsanStartProactiveRebalanceResponse `xml:"urn:vsan VsanStartProactiveRebalanceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanStartProactiveRebalanceBody) Fault() *soap.Fault { return b.Fault_ }

func VsanStartProactiveRebalance(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanStartProactiveRebalance) (*vsantypes.VsanStartProactiveRebalanceResponse, error) {
  var reqBody, resBody VsanStartProactiveRebalanceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanStopProactiveRebalanceBody struct{
    Req *vsantypes.VsanStopProactiveRebalance `xml:"urn:vsan VsanStopProactiveRebalance,omitempty"`
    Res *vsantypes.VsanStopProactiveRebalanceResponse `xml:"urn:vsan VsanStopProactiveRebalanceResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanStopProactiveRebalanceBody) Fault() *soap.Fault { return b.Fault_ }

func VsanStopProactiveRebalance(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanStopProactiveRebalance) (*vsantypes.VsanStopProactiveRebalanceResponse, error) {
  var reqBody, resBody VsanStopProactiveRebalanceBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanStopRebalanceClusterBody struct{
    Req *vsantypes.VsanStopRebalanceCluster `xml:"urn:vsan VsanStopRebalanceCluster,omitempty"`
    Res *vsantypes.VsanStopRebalanceClusterResponse `xml:"urn:vsan VsanStopRebalanceClusterResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanStopRebalanceClusterBody) Fault() *soap.Fault { return b.Fault_ }

func VsanStopRebalanceCluster(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanStopRebalanceCluster) (*vsantypes.VsanStopRebalanceClusterResponse, error) {
  var reqBody, resBody VsanStopRebalanceClusterBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVcClusterGetHclInfoBody struct{
    Req *vsantypes.VsanVcClusterGetHclInfo `xml:"urn:vsan VsanVcClusterGetHclInfo,omitempty"`
    Res *vsantypes.VsanVcClusterGetHclInfoResponse `xml:"urn:vsan VsanVcClusterGetHclInfoResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVcClusterGetHclInfoBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVcClusterGetHclInfo(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVcClusterGetHclInfo) (*vsantypes.VsanVcClusterGetHclInfoResponse, error) {
  var reqBody, resBody VsanVcClusterGetHclInfoBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVcClusterQueryVerifyHealthSystemVersionsBody struct{
    Req *vsantypes.VsanVcClusterQueryVerifyHealthSystemVersions `xml:"urn:vsan VsanVcClusterQueryVerifyHealthSystemVersions,omitempty"`
    Res *vsantypes.VsanVcClusterQueryVerifyHealthSystemVersionsResponse `xml:"urn:vsan VsanVcClusterQueryVerifyHealthSystemVersionsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVcClusterQueryVerifyHealthSystemVersionsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVcClusterQueryVerifyHealthSystemVersions(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVcClusterQueryVerifyHealthSystemVersions) (*vsantypes.VsanVcClusterQueryVerifyHealthSystemVersionsResponse, error) {
  var reqBody, resBody VsanVcClusterQueryVerifyHealthSystemVersionsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVcClusterRunVmdkLoadTestBody struct{
    Req *vsantypes.VsanVcClusterRunVmdkLoadTest `xml:"urn:vsan VsanVcClusterRunVmdkLoadTest,omitempty"`
    Res *vsantypes.VsanVcClusterRunVmdkLoadTestResponse `xml:"urn:vsan VsanVcClusterRunVmdkLoadTestResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVcClusterRunVmdkLoadTestBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVcClusterRunVmdkLoadTest(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVcClusterRunVmdkLoadTest) (*vsantypes.VsanVcClusterRunVmdkLoadTestResponse, error) {
  var reqBody, resBody VsanVcClusterRunVmdkLoadTestBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVcUpdateHclDbFromWebBody struct{
    Req *vsantypes.VsanVcUpdateHclDbFromWeb `xml:"urn:vsan VsanVcUpdateHclDbFromWeb,omitempty"`
    Res *vsantypes.VsanVcUpdateHclDbFromWebResponse `xml:"urn:vsan VsanVcUpdateHclDbFromWebResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVcUpdateHclDbFromWebBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVcUpdateHclDbFromWeb(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVcUpdateHclDbFromWeb) (*vsantypes.VsanVcUpdateHclDbFromWebResponse, error) {
  var reqBody, resBody VsanVcUpdateHclDbFromWebBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVcUploadHclDbBody struct{
    Req *vsantypes.VsanVcUploadHclDb `xml:"urn:vsan VsanVcUploadHclDb,omitempty"`
    Res *vsantypes.VsanVcUploadHclDbResponse `xml:"urn:vsan VsanVcUploadHclDbResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVcUploadHclDbBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVcUploadHclDb(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVcUploadHclDb) (*vsantypes.VsanVcUploadHclDbResponse, error) {
  var reqBody, resBody VsanVcUploadHclDbBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVcsaGetBootstrapProgressBody struct{
    Req *vsantypes.VsanVcsaGetBootstrapProgress `xml:"urn:vsan VsanVcsaGetBootstrapProgress,omitempty"`
    Res *vsantypes.VsanVcsaGetBootstrapProgressResponse `xml:"urn:vsan VsanVcsaGetBootstrapProgressResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVcsaGetBootstrapProgressBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVcsaGetBootstrapProgress(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVcsaGetBootstrapProgress) (*vsantypes.VsanVcsaGetBootstrapProgressResponse, error) {
  var reqBody, resBody VsanVcsaGetBootstrapProgressBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVdsGetMigrationPlanBody struct{
    Req *vsantypes.VsanVdsGetMigrationPlan `xml:"urn:vsan VsanVdsGetMigrationPlan,omitempty"`
    Res *vsantypes.VsanVdsGetMigrationPlanResponse `xml:"urn:vsan VsanVdsGetMigrationPlanResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVdsGetMigrationPlanBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVdsGetMigrationPlan(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVdsGetMigrationPlan) (*vsantypes.VsanVdsGetMigrationPlanResponse, error) {
  var reqBody, resBody VsanVdsGetMigrationPlanBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVdsMigrateVssBody struct{
    Req *vsantypes.VsanVdsMigrateVss `xml:"urn:vsan VsanVdsMigrateVss,omitempty"`
    Res *vsantypes.VsanVdsMigrateVssResponse `xml:"urn:vsan VsanVdsMigrateVssResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVdsMigrateVssBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVdsMigrateVss(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVdsMigrateVss) (*vsantypes.VsanVdsMigrateVssResponse, error) {
  var reqBody, resBody VsanVdsMigrateVssBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVibInstallPreflightCheckBody struct{
    Req *vsantypes.VsanVibInstallPreflightCheck `xml:"urn:vsan VsanVibInstallPreflightCheck,omitempty"`
    Res *vsantypes.VsanVibInstallPreflightCheckResponse `xml:"urn:vsan VsanVibInstallPreflightCheckResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVibInstallPreflightCheckBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVibInstallPreflightCheck(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVibInstallPreflightCheck) (*vsantypes.VsanVibInstallPreflightCheckResponse, error) {
  var reqBody, resBody VsanVibInstallPreflightCheckBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVibInstall_TaskBody struct{
    Req *vsantypes.VsanVibInstall_Task `xml:"urn:vsan VsanVibInstall_Task,omitempty"`
    Res *vsantypes.VsanVibInstall_TaskResponse `xml:"urn:vsan VsanVibInstall_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVibInstall_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVibInstall_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVibInstall_Task) (*vsantypes.VsanVibInstall_TaskResponse, error) {
  var reqBody, resBody VsanVibInstall_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVibScanBody struct{
    Req *vsantypes.VsanVibScan `xml:"urn:vsan VsanVibScan,omitempty"`
    Res *vsantypes.VsanVibScanResponse `xml:"urn:vsan VsanVibScanResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVibScanBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVibScan(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVibScan) (*vsantypes.VsanVibScanResponse, error) {
  var reqBody, resBody VsanVibScanBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitAddIscsiInitiatorGroupBody struct{
    Req *vsantypes.VsanVitAddIscsiInitiatorGroup `xml:"urn:vsan VsanVitAddIscsiInitiatorGroup,omitempty"`
    Res *vsantypes.VsanVitAddIscsiInitiatorGroupResponse `xml:"urn:vsan VsanVitAddIscsiInitiatorGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitAddIscsiInitiatorGroupBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitAddIscsiInitiatorGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitAddIscsiInitiatorGroup) (*vsantypes.VsanVitAddIscsiInitiatorGroupResponse, error) {
  var reqBody, resBody VsanVitAddIscsiInitiatorGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitAddIscsiInitiatorsToGroupBody struct{
    Req *vsantypes.VsanVitAddIscsiInitiatorsToGroup `xml:"urn:vsan VsanVitAddIscsiInitiatorsToGroup,omitempty"`
    Res *vsantypes.VsanVitAddIscsiInitiatorsToGroupResponse `xml:"urn:vsan VsanVitAddIscsiInitiatorsToGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitAddIscsiInitiatorsToGroupBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitAddIscsiInitiatorsToGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitAddIscsiInitiatorsToGroup) (*vsantypes.VsanVitAddIscsiInitiatorsToGroupResponse, error) {
  var reqBody, resBody VsanVitAddIscsiInitiatorsToGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitAddIscsiInitiatorsToTargetBody struct{
    Req *vsantypes.VsanVitAddIscsiInitiatorsToTarget `xml:"urn:vsan VsanVitAddIscsiInitiatorsToTarget,omitempty"`
    Res *vsantypes.VsanVitAddIscsiInitiatorsToTargetResponse `xml:"urn:vsan VsanVitAddIscsiInitiatorsToTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitAddIscsiInitiatorsToTargetBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitAddIscsiInitiatorsToTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitAddIscsiInitiatorsToTarget) (*vsantypes.VsanVitAddIscsiInitiatorsToTargetResponse, error) {
  var reqBody, resBody VsanVitAddIscsiInitiatorsToTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitAddIscsiLUNBody struct{
    Req *vsantypes.VsanVitAddIscsiLUN `xml:"urn:vsan VsanVitAddIscsiLUN,omitempty"`
    Res *vsantypes.VsanVitAddIscsiLUNResponse `xml:"urn:vsan VsanVitAddIscsiLUNResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitAddIscsiLUNBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitAddIscsiLUN(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitAddIscsiLUN) (*vsantypes.VsanVitAddIscsiLUNResponse, error) {
  var reqBody, resBody VsanVitAddIscsiLUNBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitAddIscsiTargetBody struct{
    Req *vsantypes.VsanVitAddIscsiTarget `xml:"urn:vsan VsanVitAddIscsiTarget,omitempty"`
    Res *vsantypes.VsanVitAddIscsiTargetResponse `xml:"urn:vsan VsanVitAddIscsiTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitAddIscsiTargetBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitAddIscsiTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitAddIscsiTarget) (*vsantypes.VsanVitAddIscsiTargetResponse, error) {
  var reqBody, resBody VsanVitAddIscsiTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitAddIscsiTargetToGroupBody struct{
    Req *vsantypes.VsanVitAddIscsiTargetToGroup `xml:"urn:vsan VsanVitAddIscsiTargetToGroup,omitempty"`
    Res *vsantypes.VsanVitAddIscsiTargetToGroupResponse `xml:"urn:vsan VsanVitAddIscsiTargetToGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitAddIscsiTargetToGroupBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitAddIscsiTargetToGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitAddIscsiTargetToGroup) (*vsantypes.VsanVitAddIscsiTargetToGroupResponse, error) {
  var reqBody, resBody VsanVitAddIscsiTargetToGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitEditIscsiLUNBody struct{
    Req *vsantypes.VsanVitEditIscsiLUN `xml:"urn:vsan VsanVitEditIscsiLUN,omitempty"`
    Res *vsantypes.VsanVitEditIscsiLUNResponse `xml:"urn:vsan VsanVitEditIscsiLUNResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitEditIscsiLUNBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitEditIscsiLUN(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitEditIscsiLUN) (*vsantypes.VsanVitEditIscsiLUNResponse, error) {
  var reqBody, resBody VsanVitEditIscsiLUNBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitEditIscsiTargetBody struct{
    Req *vsantypes.VsanVitEditIscsiTarget `xml:"urn:vsan VsanVitEditIscsiTarget,omitempty"`
    Res *vsantypes.VsanVitEditIscsiTargetResponse `xml:"urn:vsan VsanVitEditIscsiTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitEditIscsiTargetBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitEditIscsiTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitEditIscsiTarget) (*vsantypes.VsanVitEditIscsiTargetResponse, error) {
  var reqBody, resBody VsanVitEditIscsiTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitGetHomeObjectBody struct{
    Req *vsantypes.VsanVitGetHomeObject `xml:"urn:vsan VsanVitGetHomeObject,omitempty"`
    Res *vsantypes.VsanVitGetHomeObjectResponse `xml:"urn:vsan VsanVitGetHomeObjectResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitGetHomeObjectBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitGetHomeObject(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitGetHomeObject) (*vsantypes.VsanVitGetHomeObjectResponse, error) {
  var reqBody, resBody VsanVitGetHomeObjectBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitGetIscsiInitiatorGroupBody struct{
    Req *vsantypes.VsanVitGetIscsiInitiatorGroup `xml:"urn:vsan VsanVitGetIscsiInitiatorGroup,omitempty"`
    Res *vsantypes.VsanVitGetIscsiInitiatorGroupResponse `xml:"urn:vsan VsanVitGetIscsiInitiatorGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitGetIscsiInitiatorGroupBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitGetIscsiInitiatorGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitGetIscsiInitiatorGroup) (*vsantypes.VsanVitGetIscsiInitiatorGroupResponse, error) {
  var reqBody, resBody VsanVitGetIscsiInitiatorGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitGetIscsiInitiatorGroupsBody struct{
    Req *vsantypes.VsanVitGetIscsiInitiatorGroups `xml:"urn:vsan VsanVitGetIscsiInitiatorGroups,omitempty"`
    Res *vsantypes.VsanVitGetIscsiInitiatorGroupsResponse `xml:"urn:vsan VsanVitGetIscsiInitiatorGroupsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitGetIscsiInitiatorGroupsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitGetIscsiInitiatorGroups(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitGetIscsiInitiatorGroups) (*vsantypes.VsanVitGetIscsiInitiatorGroupsResponse, error) {
  var reqBody, resBody VsanVitGetIscsiInitiatorGroupsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitGetIscsiLUNBody struct{
    Req *vsantypes.VsanVitGetIscsiLUN `xml:"urn:vsan VsanVitGetIscsiLUN,omitempty"`
    Res *vsantypes.VsanVitGetIscsiLUNResponse `xml:"urn:vsan VsanVitGetIscsiLUNResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitGetIscsiLUNBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitGetIscsiLUN(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitGetIscsiLUN) (*vsantypes.VsanVitGetIscsiLUNResponse, error) {
  var reqBody, resBody VsanVitGetIscsiLUNBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitGetIscsiLUNsBody struct{
    Req *vsantypes.VsanVitGetIscsiLUNs `xml:"urn:vsan VsanVitGetIscsiLUNs,omitempty"`
    Res *vsantypes.VsanVitGetIscsiLUNsResponse `xml:"urn:vsan VsanVitGetIscsiLUNsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitGetIscsiLUNsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitGetIscsiLUNs(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitGetIscsiLUNs) (*vsantypes.VsanVitGetIscsiLUNsResponse, error) {
  var reqBody, resBody VsanVitGetIscsiLUNsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitGetIscsiTargetBody struct{
    Req *vsantypes.VsanVitGetIscsiTarget `xml:"urn:vsan VsanVitGetIscsiTarget,omitempty"`
    Res *vsantypes.VsanVitGetIscsiTargetResponse `xml:"urn:vsan VsanVitGetIscsiTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitGetIscsiTargetBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitGetIscsiTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitGetIscsiTarget) (*vsantypes.VsanVitGetIscsiTargetResponse, error) {
  var reqBody, resBody VsanVitGetIscsiTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitGetIscsiTargetsBody struct{
    Req *vsantypes.VsanVitGetIscsiTargets `xml:"urn:vsan VsanVitGetIscsiTargets,omitempty"`
    Res *vsantypes.VsanVitGetIscsiTargetsResponse `xml:"urn:vsan VsanVitGetIscsiTargetsResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitGetIscsiTargetsBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitGetIscsiTargets(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitGetIscsiTargets) (*vsantypes.VsanVitGetIscsiTargetsResponse, error) {
  var reqBody, resBody VsanVitGetIscsiTargetsBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitQueryIscsiTargetServiceVersionBody struct{
    Req *vsantypes.VsanVitQueryIscsiTargetServiceVersion `xml:"urn:vsan VsanVitQueryIscsiTargetServiceVersion,omitempty"`
    Res *vsantypes.VsanVitQueryIscsiTargetServiceVersionResponse `xml:"urn:vsan VsanVitQueryIscsiTargetServiceVersionResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitQueryIscsiTargetServiceVersionBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitQueryIscsiTargetServiceVersion(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitQueryIscsiTargetServiceVersion) (*vsantypes.VsanVitQueryIscsiTargetServiceVersionResponse, error) {
  var reqBody, resBody VsanVitQueryIscsiTargetServiceVersionBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitRemoveIscsiInitiatorGroupBody struct{
    Req *vsantypes.VsanVitRemoveIscsiInitiatorGroup `xml:"urn:vsan VsanVitRemoveIscsiInitiatorGroup,omitempty"`
    Res *vsantypes.VsanVitRemoveIscsiInitiatorGroupResponse `xml:"urn:vsan VsanVitRemoveIscsiInitiatorGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitRemoveIscsiInitiatorGroupBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitRemoveIscsiInitiatorGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitRemoveIscsiInitiatorGroup) (*vsantypes.VsanVitRemoveIscsiInitiatorGroupResponse, error) {
  var reqBody, resBody VsanVitRemoveIscsiInitiatorGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitRemoveIscsiInitiatorsFromGroupBody struct{
    Req *vsantypes.VsanVitRemoveIscsiInitiatorsFromGroup `xml:"urn:vsan VsanVitRemoveIscsiInitiatorsFromGroup,omitempty"`
    Res *vsantypes.VsanVitRemoveIscsiInitiatorsFromGroupResponse `xml:"urn:vsan VsanVitRemoveIscsiInitiatorsFromGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitRemoveIscsiInitiatorsFromGroupBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitRemoveIscsiInitiatorsFromGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitRemoveIscsiInitiatorsFromGroup) (*vsantypes.VsanVitRemoveIscsiInitiatorsFromGroupResponse, error) {
  var reqBody, resBody VsanVitRemoveIscsiInitiatorsFromGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitRemoveIscsiInitiatorsFromTargetBody struct{
    Req *vsantypes.VsanVitRemoveIscsiInitiatorsFromTarget `xml:"urn:vsan VsanVitRemoveIscsiInitiatorsFromTarget,omitempty"`
    Res *vsantypes.VsanVitRemoveIscsiInitiatorsFromTargetResponse `xml:"urn:vsan VsanVitRemoveIscsiInitiatorsFromTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitRemoveIscsiInitiatorsFromTargetBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitRemoveIscsiInitiatorsFromTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitRemoveIscsiInitiatorsFromTarget) (*vsantypes.VsanVitRemoveIscsiInitiatorsFromTargetResponse, error) {
  var reqBody, resBody VsanVitRemoveIscsiInitiatorsFromTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitRemoveIscsiLUNBody struct{
    Req *vsantypes.VsanVitRemoveIscsiLUN `xml:"urn:vsan VsanVitRemoveIscsiLUN,omitempty"`
    Res *vsantypes.VsanVitRemoveIscsiLUNResponse `xml:"urn:vsan VsanVitRemoveIscsiLUNResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitRemoveIscsiLUNBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitRemoveIscsiLUN(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitRemoveIscsiLUN) (*vsantypes.VsanVitRemoveIscsiLUNResponse, error) {
  var reqBody, resBody VsanVitRemoveIscsiLUNBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitRemoveIscsiTargetBody struct{
    Req *vsantypes.VsanVitRemoveIscsiTarget `xml:"urn:vsan VsanVitRemoveIscsiTarget,omitempty"`
    Res *vsantypes.VsanVitRemoveIscsiTargetResponse `xml:"urn:vsan VsanVitRemoveIscsiTargetResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitRemoveIscsiTargetBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitRemoveIscsiTarget(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitRemoveIscsiTarget) (*vsantypes.VsanVitRemoveIscsiTargetResponse, error) {
  var reqBody, resBody VsanVitRemoveIscsiTargetBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanVitRemoveIscsiTargetFromGroupBody struct{
    Req *vsantypes.VsanVitRemoveIscsiTargetFromGroup `xml:"urn:vsan VsanVitRemoveIscsiTargetFromGroup,omitempty"`
    Res *vsantypes.VsanVitRemoveIscsiTargetFromGroupResponse `xml:"urn:vsan VsanVitRemoveIscsiTargetFromGroupResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanVitRemoveIscsiTargetFromGroupBody) Fault() *soap.Fault { return b.Fault_ }

func VsanVitRemoveIscsiTargetFromGroup(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanVitRemoveIscsiTargetFromGroup) (*vsantypes.VsanVitRemoveIscsiTargetFromGroupResponse, error) {
  var reqBody, resBody VsanVitRemoveIscsiTargetFromGroupBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type VsanWaitForVsanHealthGenerationIdChangeBody struct{
    Req *vsantypes.VsanWaitForVsanHealthGenerationIdChange `xml:"urn:vsan VsanWaitForVsanHealthGenerationIdChange,omitempty"`
    Res *vsantypes.VsanWaitForVsanHealthGenerationIdChangeResponse `xml:"urn:vsan VsanWaitForVsanHealthGenerationIdChangeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *VsanWaitForVsanHealthGenerationIdChangeBody) Fault() *soap.Fault { return b.Fault_ }

func VsanWaitForVsanHealthGenerationIdChange(ctx context.Context, r soap.RoundTripper, req *vsantypes.VsanWaitForVsanHealthGenerationIdChange) (*vsantypes.VsanWaitForVsanHealthGenerationIdChangeResponse, error) {
  var reqBody, resBody VsanWaitForVsanHealthGenerationIdChangeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type WaitForUpdatesBody struct{
    Req *vsantypes.WaitForUpdates `xml:"urn:vsan WaitForUpdates,omitempty"`
    Res *vsantypes.WaitForUpdatesResponse `xml:"urn:vsan WaitForUpdatesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *WaitForUpdatesBody) Fault() *soap.Fault { return b.Fault_ }

func WaitForUpdates(ctx context.Context, r soap.RoundTripper, req *vsantypes.WaitForUpdates) (*vsantypes.WaitForUpdatesResponse, error) {
  var reqBody, resBody WaitForUpdatesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type WaitForUpdatesExBody struct{
    Req *vsantypes.WaitForUpdatesEx `xml:"urn:vsan WaitForUpdatesEx,omitempty"`
    Res *vsantypes.WaitForUpdatesExResponse `xml:"urn:vsan WaitForUpdatesExResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *WaitForUpdatesExBody) Fault() *soap.Fault { return b.Fault_ }

func WaitForUpdatesEx(ctx context.Context, r soap.RoundTripper, req *vsantypes.WaitForUpdatesEx) (*vsantypes.WaitForUpdatesExResponse, error) {
  var reqBody, resBody WaitForUpdatesExBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type XmlToCustomizationSpecItemBody struct{
    Req *vsantypes.XmlToCustomizationSpecItem `xml:"urn:vsan XmlToCustomizationSpecItem,omitempty"`
    Res *vsantypes.XmlToCustomizationSpecItemResponse `xml:"urn:vsan XmlToCustomizationSpecItemResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *XmlToCustomizationSpecItemBody) Fault() *soap.Fault { return b.Fault_ }

func XmlToCustomizationSpecItem(ctx context.Context, r soap.RoundTripper, req *vsantypes.XmlToCustomizationSpecItem) (*vsantypes.XmlToCustomizationSpecItemResponse, error) {
  var reqBody, resBody XmlToCustomizationSpecItemBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ZeroFillVirtualDisk_TaskBody struct{
    Req *vsantypes.ZeroFillVirtualDisk_Task `xml:"urn:vsan ZeroFillVirtualDisk_Task,omitempty"`
    Res *vsantypes.ZeroFillVirtualDisk_TaskResponse `xml:"urn:vsan ZeroFillVirtualDisk_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ZeroFillVirtualDisk_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ZeroFillVirtualDisk_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ZeroFillVirtualDisk_Task) (*vsantypes.ZeroFillVirtualDisk_TaskResponse, error) {
  var reqBody, resBody ZeroFillVirtualDisk_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ConfigureVcha_TaskBody struct{
    Req *vsantypes.ConfigureVcha_Task `xml:"urn:vsan configureVcha_Task,omitempty"`
    Res *vsantypes.ConfigureVcha_TaskResponse `xml:"urn:vsan configureVcha_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ConfigureVcha_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ConfigureVcha_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ConfigureVcha_Task) (*vsantypes.ConfigureVcha_TaskResponse, error) {
  var reqBody, resBody ConfigureVcha_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreatePassiveNode_TaskBody struct{
    Req *vsantypes.CreatePassiveNode_Task `xml:"urn:vsan createPassiveNode_Task,omitempty"`
    Res *vsantypes.CreatePassiveNode_TaskResponse `xml:"urn:vsan createPassiveNode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreatePassiveNode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreatePassiveNode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreatePassiveNode_Task) (*vsantypes.CreatePassiveNode_TaskResponse, error) {
  var reqBody, resBody CreatePassiveNode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type CreateWitnessNode_TaskBody struct{
    Req *vsantypes.CreateWitnessNode_Task `xml:"urn:vsan createWitnessNode_Task,omitempty"`
    Res *vsantypes.CreateWitnessNode_TaskResponse `xml:"urn:vsan createWitnessNode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *CreateWitnessNode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func CreateWitnessNode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.CreateWitnessNode_Task) (*vsantypes.CreateWitnessNode_TaskResponse, error) {
  var reqBody, resBody CreateWitnessNode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DeployVcha_TaskBody struct{
    Req *vsantypes.DeployVcha_Task `xml:"urn:vsan deployVcha_Task,omitempty"`
    Res *vsantypes.DeployVcha_TaskResponse `xml:"urn:vsan deployVcha_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DeployVcha_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DeployVcha_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DeployVcha_Task) (*vsantypes.DeployVcha_TaskResponse, error) {
  var reqBody, resBody DeployVcha_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type DestroyVcha_TaskBody struct{
    Req *vsantypes.DestroyVcha_Task `xml:"urn:vsan destroyVcha_Task,omitempty"`
    Res *vsantypes.DestroyVcha_TaskResponse `xml:"urn:vsan destroyVcha_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *DestroyVcha_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func DestroyVcha_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.DestroyVcha_Task) (*vsantypes.DestroyVcha_TaskResponse, error) {
  var reqBody, resBody DestroyVcha_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type FetchSoftwarePackagesBody struct{
    Req *vsantypes.FetchSoftwarePackages `xml:"urn:vsan fetchSoftwarePackages,omitempty"`
    Res *vsantypes.FetchSoftwarePackagesResponse `xml:"urn:vsan fetchSoftwarePackagesResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *FetchSoftwarePackagesBody) Fault() *soap.Fault { return b.Fault_ }

func FetchSoftwarePackages(ctx context.Context, r soap.RoundTripper, req *vsantypes.FetchSoftwarePackages) (*vsantypes.FetchSoftwarePackagesResponse, error) {
  var reqBody, resBody FetchSoftwarePackagesBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetClusterModeBody struct{
    Req *vsantypes.GetClusterMode `xml:"urn:vsan getClusterMode,omitempty"`
    Res *vsantypes.GetClusterModeResponse `xml:"urn:vsan getClusterModeResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetClusterModeBody) Fault() *soap.Fault { return b.Fault_ }

func GetClusterMode(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetClusterMode) (*vsantypes.GetClusterModeResponse, error) {
  var reqBody, resBody GetClusterModeBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type GetVchaConfigBody struct{
    Req *vsantypes.GetVchaConfig `xml:"urn:vsan getVchaConfig,omitempty"`
    Res *vsantypes.GetVchaConfigResponse `xml:"urn:vsan getVchaConfigResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *GetVchaConfigBody) Fault() *soap.Fault { return b.Fault_ }

func GetVchaConfig(ctx context.Context, r soap.RoundTripper, req *vsantypes.GetVchaConfig) (*vsantypes.GetVchaConfigResponse, error) {
  var reqBody, resBody GetVchaConfigBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InitiateFailover_TaskBody struct{
    Req *vsantypes.InitiateFailover_Task `xml:"urn:vsan initiateFailover_Task,omitempty"`
    Res *vsantypes.InitiateFailover_TaskResponse `xml:"urn:vsan initiateFailover_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InitiateFailover_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func InitiateFailover_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.InitiateFailover_Task) (*vsantypes.InitiateFailover_TaskResponse, error) {
  var reqBody, resBody InitiateFailover_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type InstallDateBody struct{
    Req *vsantypes.InstallDate `xml:"urn:vsan installDate,omitempty"`
    Res *vsantypes.InstallDateResponse `xml:"urn:vsan installDateResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *InstallDateBody) Fault() *soap.Fault { return b.Fault_ }

func InstallDate(ctx context.Context, r soap.RoundTripper, req *vsantypes.InstallDate) (*vsantypes.InstallDateResponse, error) {
  var reqBody, resBody InstallDateBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type PrepareVcha_TaskBody struct{
    Req *vsantypes.PrepareVcha_Task `xml:"urn:vsan prepareVcha_Task,omitempty"`
    Res *vsantypes.PrepareVcha_TaskResponse `xml:"urn:vsan prepareVcha_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *PrepareVcha_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func PrepareVcha_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.PrepareVcha_Task) (*vsantypes.PrepareVcha_TaskResponse, error) {
  var reqBody, resBody PrepareVcha_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type QueryDatacenterConfigOptionDescriptorBody struct{
    Req *vsantypes.QueryDatacenterConfigOptionDescriptor `xml:"urn:vsan queryDatacenterConfigOptionDescriptor,omitempty"`
    Res *vsantypes.QueryDatacenterConfigOptionDescriptorResponse `xml:"urn:vsan queryDatacenterConfigOptionDescriptorResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *QueryDatacenterConfigOptionDescriptorBody) Fault() *soap.Fault { return b.Fault_ }

func QueryDatacenterConfigOptionDescriptor(ctx context.Context, r soap.RoundTripper, req *vsantypes.QueryDatacenterConfigOptionDescriptor) (*vsantypes.QueryDatacenterConfigOptionDescriptorResponse, error) {
  var reqBody, resBody QueryDatacenterConfigOptionDescriptorBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type ReloadVirtualMachineFromPath_TaskBody struct{
    Req *vsantypes.ReloadVirtualMachineFromPath_Task `xml:"urn:vsan reloadVirtualMachineFromPath_Task,omitempty"`
    Res *vsantypes.ReloadVirtualMachineFromPath_TaskResponse `xml:"urn:vsan reloadVirtualMachineFromPath_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *ReloadVirtualMachineFromPath_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func ReloadVirtualMachineFromPath_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.ReloadVirtualMachineFromPath_Task) (*vsantypes.ReloadVirtualMachineFromPath_TaskResponse, error) {
  var reqBody, resBody ReloadVirtualMachineFromPath_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetClusterMode_TaskBody struct{
    Req *vsantypes.SetClusterMode_Task `xml:"urn:vsan setClusterMode_Task,omitempty"`
    Res *vsantypes.SetClusterMode_TaskResponse `xml:"urn:vsan setClusterMode_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetClusterMode_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func SetClusterMode_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetClusterMode_Task) (*vsantypes.SetClusterMode_TaskResponse, error) {
  var reqBody, resBody SetClusterMode_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type SetCustomValueBody struct{
    Req *vsantypes.SetCustomValue `xml:"urn:vsan setCustomValue,omitempty"`
    Res *vsantypes.SetCustomValueResponse `xml:"urn:vsan setCustomValueResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *SetCustomValueBody) Fault() *soap.Fault { return b.Fault_ }

func SetCustomValue(ctx context.Context, r soap.RoundTripper, req *vsantypes.SetCustomValue) (*vsantypes.SetCustomValueResponse, error) {
  var reqBody, resBody SetCustomValueBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

  type UnregisterVApp_TaskBody struct{
    Req *vsantypes.UnregisterVApp_Task `xml:"urn:vsan unregisterVApp_Task,omitempty"`
    Res *vsantypes.UnregisterVApp_TaskResponse `xml:"urn:vsan unregisterVApp_TaskResponse,omitempty"`
    Fault_ *soap.Fault `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
  }

  func (b *UnregisterVApp_TaskBody) Fault() *soap.Fault { return b.Fault_ }

func UnregisterVApp_Task(ctx context.Context, r soap.RoundTripper, req *vsantypes.UnregisterVApp_Task) (*vsantypes.UnregisterVApp_TaskResponse, error) {
  var reqBody, resBody UnregisterVApp_TaskBody

  reqBody.Req = req

  if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
    return nil, err
  }

  return resBody.Res, nil
}

