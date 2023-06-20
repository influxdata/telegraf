#include "ublox-utils/include/ublox_reader.h"
#include "ublox-utils/include/ubx.h"
#include <cstdlib>
#include <stdbool.h>
#include <string.h>

extern "C" {
void *ublox_reader_new() { return new UbloxReader; }

void ublox_reader_free(void *reader) { delete (UbloxReader *)reader; }

// param err readable error in case of failure, note you must free it
bool ublox_reader_init(void *reader, const char *device, char **err) {
  UbloxReader *ublox_reader = (UbloxReader *)reader;
  std::string serr;
  bool retval = ublox_reader->init(device, &serr);
  if (retval == false && err) {
    *err = (char *)malloc(serr.size());
    strcpy(*err, serr.c_str());
  }
  return retval;
}

void ublox_reader_close(void *reader) { ((UbloxReader *)reader)->close(); }

// note you must free err if error happened
int ublox_reader_read(void *reader, bool *is_active, double *lat, double *lon,
                      double *heading, unsigned int *pdop,
                      unsigned int *fusion_mode, long long *sec, long long *nsec, bool wait_for_data,
                      char **err) {
  UbloxReader *ublox_reader = (UbloxReader *)reader;

  const void *msg = nullptr;
  size_t len = 0;
  std::string serr;
  for (;;) {
    switch (ublox_reader->get(&msg, &len, wait_for_data, &serr)) {
    case UbloxReader::Status::None:
      return 0;
    case UbloxReader::Closed:
      return 0;
    case UbloxReader::InvalidMessage:
    case UbloxReader::NMEAMessage:
      // do nothing
      break;
    case UbloxReader::UBXMessage:
      if (ubx::messageCId(msg, len) == ubx::NAV_PVT) {
        const ubx::NavPvt *nav_pvt = (const ubx::NavPvt *)msg;
        if (is_active) {
          *is_active = nav_pvt->payload.flags & ubx::GNSSFixOk;
        }
        if (lat) {
          *lat = nav_pvt->payload.lat / 10'000'000.;
        }
        if (lon) {
          *lon = nav_pvt->payload.lon / 10'000'000.;
        }
        if (heading) {
          *heading = nav_pvt->payload.headVeh / 100'000.;
        }
        if (pdop) {
          *pdop = nav_pvt->payload.pDOP;
        }
        if (sec) {
          *sec = ubx::getUtcSec(nav_pvt);
        }
        if (nsec) {
          *nsec = nav_pvt->payload.nano;
        }

        return 1;
      } else if (ubx::messageCId(msg, len) == ubx::ESF_STATUS) {
        const ubx::EsfStatus *esf_status = (const ubx::EsfStatus *)msg;
        if (fusion_mode) {
          *fusion_mode = esf_status->payload.fusionMode;
        }
        // do not return here, wait for NAV_PVT message
      }
      break;
    case UbloxReader::Error:
      if (err) {
        *err = (char *)malloc(serr.size());
        strcpy(*err, serr.c_str());
      }
      return -1;
    }
  }
}
}
