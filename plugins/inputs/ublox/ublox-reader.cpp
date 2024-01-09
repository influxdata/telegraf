#include "ublox-utils/include/ublox_reader.h"
#include "ublox-utils/include/ubx.h"
#include <cstdlib>
#include <stdbool.h>
#include <string.h>


#define TRY_SET(ptr, val) \
  if (ptr) {              \
    *ptr = val;           \
  }


extern "C" {
void *ublox_reader_new() {
  return new UbloxReader;
}

void ublox_reader_free(void *reader) {
  delete (UbloxReader *)reader;
}

// param err readable error in case of failure, note you must free it
bool ublox_reader_init(void *reader, const char *device, char **err) {
  UbloxReader *ublox_reader = (UbloxReader *)reader;
  std::string  serr;
  bool         retval = ublox_reader->init(device, &serr);
  if (retval == false && err) {
    *err = (char *)malloc(serr.size() + 1 /*\0*/);
    strcpy(*err, serr.c_str());
  }
  return retval;
}

void ublox_reader_close(void *reader) {
  ((UbloxReader *)reader)->close();
}

// note you must free err if error happened
int ublox_reader_read(void         *reader,
                      bool         *is_active,
                      double       *lat,
                      double       *lon,
                      double       *horizontal_acc,
                      double       *heading,
                      double       *heading_of_mot,
                      double       *heading_acc,
                      bool         *heading_is_valid,
                      double       *speed,
                      double       *speed_acc,
                      unsigned int *pdop,
                      unsigned int *sat_num,
                      unsigned int *fix_type,
                      unsigned int *fusion_mode,
                      char          sensor_arr[4 * 16],
                      unsigned int *sensors_count,
                      long long    *sec,
                      long long    *nsec,
                      bool          wait_for_data,
                      char        **err) {
  UbloxReader *ublox_reader = (UbloxReader *)reader;

  const void *msg = nullptr;
  size_t      len = 0;
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
        TRY_SET(is_active, nav_pvt->payload.flags & ubx::GNSSFixOk);

        TRY_SET(lat, nav_pvt->payload.lat / 10'000'000.);
        TRY_SET(lon, nav_pvt->payload.lon / 10'000'000.);

        TRY_SET(horizontal_acc, nav_pvt->payload.hAcc / 1'000.);

        TRY_SET(heading, nav_pvt->payload.headVeh / 100'000.);
        TRY_SET(heading_of_mot, nav_pvt->payload.headMot / 100'000.);
        TRY_SET(heading_acc, nav_pvt->payload.headAcc / 100'000.);
        TRY_SET(heading_is_valid,
                nav_pvt->payload.flags & ubx::Flags::HeadVehValid);

        TRY_SET(speed, nav_pvt->payload.gSpeed / 1'000.);
        TRY_SET(speed_acc, nav_pvt->payload.sAcc / 1'000.);

        TRY_SET(pdop, nav_pvt->payload.pDOP);
        TRY_SET(sat_num, nav_pvt->payload.numSV);
        TRY_SET(fix_type, nav_pvt->payload.fixType);

        TRY_SET(sec, ubx::getUtcSec(nav_pvt));
        TRY_SET(nsec, nav_pvt->payload.nano);

        return 1;
      } else if (ubx::messageCId(msg, len) == ubx::ESF_STATUS) {
        const ubx::EsfStatus *esf_status = (const ubx::EsfStatus *)msg;

        TRY_SET(fusion_mode, esf_status->payload.fusionMode);

        if (esf_status->payload.numSens <= 16) {
          TRY_SET(sensors_count, esf_status->payload.numSens);
          memcpy(sensor_arr,
                 esf_status->payload.sensors,
                 4 * esf_status->payload.numSens);
        }

        // do not return here, wait for NAV_PVT message
      }
      break;
    case UbloxReader::Error:
      if (err) {
        *err = (char *)malloc(serr.size() + 1 /*\0*/);
        strcpy(*err, serr.c_str());
      }
      return -1;
    }
  }
}
}
