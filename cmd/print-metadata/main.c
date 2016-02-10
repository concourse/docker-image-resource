#include <stdio.h>
#include <unistd.h>
#include <string.h>

#include <jansson.h>

#include "pwd_ng.h"

extern char **environ;

int main() {
  struct passwd *pw = getpwuid(getuid());
  if (pw == NULL) {
    perror("failed to get username: ");
    return 1;
  }

  json_t *root = json_object();

  if (json_object_set_new(root, "user", json_string(pw->pw_name)) != 0) {
    printf("failed to set user value: %s", pw->pw_name);
    return 1;
  }

  json_t *env_json = json_array();
  for (char** env = environ; *env != NULL; env++) {
    if (strncmp(*env, "HOSTNAME=", strlen("HOSTNAME=")) != 0) {
      json_array_append_new(env_json, json_string(*env));
    }
  }

  if (json_object_set_new(root, "env", env_json) != 0) {
    printf("failed to set env value");
    return 1;
  }

  char* output = json_dumps(root, JSON_INDENT(2));
  printf("%s\n", output);

  free(output);
  json_decref(root);

  return 0;
}
