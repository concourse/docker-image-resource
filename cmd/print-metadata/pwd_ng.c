#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <sys/types.h>

#include "pwd_ng.h"

#define _GETPWNAM_NEXT(x, y)                  \
  do {                                        \
      if ((y) != NULL) {                      \
        (x) = (y) + 1;                        \
      }                                       \
                                              \
      (y) = strchr((x), ':');                 \
                                              \
      /* Search for \n in last iteration */   \
      if ((y) == NULL) {                      \
        (y) = strchr((x), '\n');              \
      }                                       \
                                              \
      if ((y) == NULL) {                      \
        goto done;                            \
      }                                       \
                                              \
      *(y) = '\0';                            \
  } while(0);

/* Instead of using getpwuid from glibc, the following custom version is used
 * because we need to bypass dynamically loading the nsswitch libraries.
 * The version of glibc inside a container may be different than the version
 * that wshd is compiled for, leading to undefined behavior. */
struct passwd *getpwuid(uid_t uid) {
  static struct passwd passwd;
  static char buf[1024];
  struct passwd *_passwd = NULL;
  FILE *f;
  char *p, *q;

  char uidstr[128];
  if (sprintf(uidstr, "%d", uid) == 0) {
    perror("failed to format uid");
    goto done;
  }

  f = fopen("/etc/passwd", "r");
  if (f == NULL) {
    goto done;
  }

  while (fgets(buf, sizeof(buf), f) != NULL) {
    p = buf;
    q = NULL;

    /* Username */
    _GETPWNAM_NEXT(p, q);
    passwd.pw_name = p;

    /* User password */
    _GETPWNAM_NEXT(p, q);
    passwd.pw_passwd = p;

    /* User ID */
    _GETPWNAM_NEXT(p, q);
    passwd.pw_uid = atoi(p);

    if (strcmp(p, uidstr) != 0) {
      continue;
    }

    /* Group ID */
    _GETPWNAM_NEXT(p, q);
    passwd.pw_gid = atoi(p);

    /* User information */
    _GETPWNAM_NEXT(p, q);
    passwd.pw_gecos = p;

    /* Home directory */
    _GETPWNAM_NEXT(p, q);
    passwd.pw_dir = p;

    /* Shell program */
    _GETPWNAM_NEXT(p, q);
    passwd.pw_shell = p;

    /* Done! */
    _passwd = &passwd;
    goto done;
  }

done:
  if (f != NULL) {
    fclose(f);
  }

  return _passwd;
}

#undef _GETPWNAM_NEXT
