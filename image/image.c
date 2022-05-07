#include "image.h"

#include <vips/vips.h>

int manipulate(void *buf, size_t len, Result *result, Options options) {
  VipsImage *orig = vips_image_new_from_buffer(buf, len, "", "access",
                                               VIPS_ACCESS_SEQUENTIAL, NULL);

  int width = options.width;
  int height = options.height;
  if (width == 0) {
    width = ((float)height / orig->Ysize) * orig->Xsize;
  }
  if (height == 0) {
    height = ((float)width / orig->Xsize) * orig->Ysize;
  }

  VipsImage *out = vips_image_new();
  int err = vips_thumbnail_image(orig, &out, width, "height", height, "crop",
                                 options.crop, "size", options.size, NULL);
  if (err != 0) {
    return err;
  }

  g_object_unref(orig);

  if (options.blur > 0) {
    VipsImage *blurred = vips_image_new();
    vips_gaussblur(out, &blurred, options.blur, NULL);
    g_object_unref(out);
    out = blurred;
  }

  if (options.quality == 0) {
    options.quality = 85;
  }

  switch (options.format) {
  case JPEG:
    err = vips_jpegsave_buffer(out, &result->buf, &result->len, "Q",
                               options.quality, "strip", TRUE, NULL);
    break;
  case PNG:
    err = vips_pngsave_buffer(out, &result->buf, &result->len, "Q",
                              options.quality, NULL);
    break;
  case WEBP:
    err = vips_webpsave_buffer(out, &result->buf, &result->len, "Q",
                               options.quality, NULL);
    break;
  default:
    err = 1;
  }

  g_object_unref(out);

  return err;
}
