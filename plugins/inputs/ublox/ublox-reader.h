#include <stdbool.h>
#include <stdlib.h>


void *ublox_reader_new();
void  ublox_reader_free(void *reader);
bool  ublox_reader_init(void *reader, const char *device, char **err);
void  ublox_reader_close(void *reader);

/**\param sensor_arr 4 bytes of information per each sensor
 * \return -1 in case of error, 0 if no data or closed, 1 if data was obtained
 * \note there is no guarantee that fields: fusion_mode, sensor*, *_version and
 * hdop will be set per each call of the function
 */
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
                      char        **err);

/**\brief try to update sw, hw and fw versions
 * return -1 in case of error, otherwise 0
 */
int ublox_reader_update_version_info(void *reader, char **err);
