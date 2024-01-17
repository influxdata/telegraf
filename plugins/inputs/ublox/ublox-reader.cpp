#include "ublox-utils/include/ublox_config_protocol.h"
#include "ublox-utils/include/ublox_reader.h"
#include "ublox-utils/include/ubx.h"
#include <cstdlib>
#include <stdbool.h>
#include <string.h>
#include <unistd.h>


#define FW_PREFIX "FWVER="


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
                      char          sw_version[30],
                      char          hw_version[30],
                      char          fw_version[30],
                      unsigned int *hdop,
                      long long    *sec,
                      long long    *nsec,
                      bool          wait_for_data,
                      char        **err) {
  UbloxReader *ublox_reader = (UbloxReader *)reader;

  const void *msg = nullptr;
  size_t      len = 0;
  std::string serr;
  for (;;) {
    // return only at NAV-PVT message
    switch (ublox_reader->pop(&msg, &len, wait_for_data, &serr)) {
    case UbloxReader::Status::None:
      return 0;
    case UbloxReader::Closed:
      return 0;
    case UbloxReader::InvalidMessage:
    case UbloxReader::NMEAMessage:
      // do nothing
      break;
    case UbloxReader::UBXMessage:
      switch (ubx::messageCId(msg, len)) {
      case ubx::NAV_PVT: {
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
      }
      case ubx::ESF_STATUS: {
        const ubx::EsfStatus *esf_status = (const ubx::EsfStatus *)msg;

        TRY_SET(fusion_mode, esf_status->payload.fusionMode);

        if (esf_status->payload.numSens <= 16) {
          TRY_SET(sensors_count, esf_status->payload.numSens);
          memcpy(sensor_arr,
                 esf_status->payload.sensors,
                 4 * esf_status->payload.numSens);
        }
      } break;
      case ubx::MON_VER: {
        const ubx::MonVer *mon_ver = (const ubx::MonVer *)msg;
        strcpy(sw_version, mon_ver->payload.swVersion);
        strcpy(hw_version, mon_ver->payload.hwVersion);
        for (int i = 0;
             mon_ver->header.len >
             sizeof(ubx::MonVerPayload) + i * sizeof(ubx::MonVerRepeatedGroup);
             ++i) {
          if (strncmp(mon_ver->payload.extensions[i].extension,
                      FW_PREFIX,
                      strlen(FW_PREFIX)) == 0) {
            strcpy(fw_version,
                   mon_ver->payload.extensions[i].extension +
                       strlen(FW_PREFIX));
            break;
          }
        }
      } break;
      case ubx::NAV_DOP: {
        const ubx::NavDop *nav_dop = (const ubx::NavDop *)msg;
        TRY_SET(hdop, nav_dop->payload.hDOP);
      } break;
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


/**\brief try to update sw, hw and fw versions
 */
int ublox_reader_update_version_info(void *reader, char **err) {
  // FIXME currently we can not write anything with using UbloxReader
  UbloxReader *ublox_reader = (UbloxReader *)reader;

  uint8_t buf[256];
  size_t  len = makeConfigRequest(-1, buf);


  std::string serr;
  if (ublox_reader->push(buf, len, &serr) == false) {
    if (err) {
      *err = (char *)malloc(serr.size() + 1 /*\0*/);
      strcpy(*err, serr.c_str());
    }
    return -1;
  }

  return 0;
}
}
