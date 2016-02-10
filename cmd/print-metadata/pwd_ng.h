#ifndef PWD_H
#define PWD_H

#include <stdint.h>
#include <sys/types.h>

struct passwd {
  char *pw_name;   /* Username. */
  char *pw_passwd; /* Password. */
  uint32_t pw_uid; /* User ID. */
  uint32_t pw_gid; /* Group ID. */
  char *pw_gecos;  /* Real name. */
  char *pw_dir;    /* Home directory. */
  char *pw_shell;  /* Shell program. */
};

struct passwd *getpwuid(uid_t uid);

#endif
