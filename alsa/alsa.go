package alsa

/*
#cgo CFLAGS: -lasound
#cgo CFLAGS: -I/usr/include/alsa  
#include <unistd.h>
#include <fcntl.h>

#include <alsa/asoundlib.h>
#include <alsa/mixer.h>

typedef enum {
    AUDIO_VOLUME_SET,
    AUDIO_VOLUME_GET,
} audio_volume_action;

int audio_volume(audio_volume_action action, long* outvol)
{
	int ret = 0;
    snd_mixer_t* handle;
    snd_mixer_elem_t* elem;
    snd_mixer_selem_id_t* sid;

    static const char* mix_name = "Master";
    static const char* card = "default";
    static int mix_index = 0;


    snd_mixer_selem_id_alloca(&sid);

    //sets simple-mixer index and name
    snd_mixer_selem_id_set_index(sid, mix_index);
    snd_mixer_selem_id_set_name(sid, mix_name);

        if ((snd_mixer_open(&handle, 0)) < 0)
        return -1;
    if ((snd_mixer_attach(handle, card)) < 0) {
        snd_mixer_close(handle);
        return -2;
    }
    if ((snd_mixer_selem_register(handle, NULL, NULL)) < 0) {
        snd_mixer_close(handle);
        return -3;
    }
    ret = snd_mixer_load(handle);
    if (ret < 0) {
        snd_mixer_close(handle);
        return -4;
    }
    elem = snd_mixer_find_selem(handle, sid);
    if (!elem) {
        snd_mixer_close(handle);
        return -5;
    }

    long minv, maxv;

    snd_mixer_selem_get_playback_volume_range (elem, &minv, &maxv);

    if(action == AUDIO_VOLUME_GET) {
        if(snd_mixer_selem_get_playback_volume(elem, 0, outvol) < 0) {
            snd_mixer_close(handle);
            return -6;
        }

        // make the value bound to 100
        *outvol -= minv;
        maxv -= minv;
        minv = 0;
        *outvol = 100 * (*outvol) / maxv; // make the value bound from 0 to 100
    }
    else if(action == AUDIO_VOLUME_SET) {
        if(*outvol < 0 || *outvol > 100) // out of bounds
            return -7;
        *outvol = (*outvol * (maxv - minv) / (100-1)) + minv;

        if(snd_mixer_selem_set_playback_volume(elem, 0, *outvol) < 0) {
            snd_mixer_close(handle);
            return -8;
        }
        if(snd_mixer_selem_set_playback_volume(elem, 1, *outvol) < 0) {
            snd_mixer_close(handle);
            return -9;
        }
    }

    snd_mixer_close(handle);
    return 0;
}

long GetVol(){
    long vol = -1;
    return audio_volume(AUDIO_VOLUME_GET, &vol);
}

void SetVol(long volume){
    audio_volume(AUDIO_VOLUME_SET, &volume);

}
//int main(void)
//{
//    long vol = -1;
//    printf("Ret %i\n", audio_volume(AUDIO_VOLUME_GET, &vol));
//    printf("Master volume is %i\n", vol);
//
//    vol = 100;
//    printf("Ret %i\n", audio_volume(AUDIO_VOLUME_SET, &vol));
//
//    return 0;
//}
*/
import "C"

import "errors"
func GetVolume() (uint8, error){
    if val := uint8(C.GetVol()); val== -1{
        return -1, errors.New("Unable to get volume.")
    }
    return val
}
func SetVolume(vol uint8){
    if vol >100 || vol <0{
        panic("Volume exceeds acceptable boundaries (0-100).")

    } else {
        C.SetVol(C.long(vol))
    }

}
