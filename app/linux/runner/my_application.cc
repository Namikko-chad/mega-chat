#include "my_application.h"

#include <flutter_linux/flutter_linux.h>
#ifdef GDK_WINDOWING_X11
#include <gdk/gdkx.h>
#endif

#include "flutter/generated_plugin_registrant.h"

struct _MyApplication {
  GtkApplication parent_instance;
  char** dart_entrypoint_arguments;
  FlMethodChannel* files;
};

G_DEFINE_TYPE(MyApplication, my_application, GTK_TYPE_APPLICATION)

// Called when first Flutter frame received.
static void first_frame_cb(MyApplication* self, FlView* view) {
  gtk_widget_show(gtk_widget_get_toplevel(GTK_WIDGET(view)));
}

// "dev.megachat/files" "clipboard" (lib/data/local_files.dart): files
// copied in a file manager as {"uris": [...]}, or a picture as {"png":
// bytes}; nothing when the clipboard has only text, or text with a picture
// (spreadsheet cells), so that a paste stays a text paste. Read with the
// asynchronous requests: the call is answered from their callbacks.
static void clipboard_respond(gpointer call, FlValue* value) {
  fl_method_call_respond_success(FL_METHOD_CALL(call), value, nullptr);
  g_object_unref(call);
}

static void clipboard_empty(gpointer call) {
  g_autoptr(FlValue) out = fl_value_new_map();
  clipboard_respond(call, out);
}

static void clipboard_got_image(GtkClipboard* cb, GdkPixbuf* pixbuf,
                                gpointer call) {
  g_autoptr(FlValue) out = fl_value_new_map();
  gchar* buf = nullptr;
  gsize len = 0;
  if (pixbuf != nullptr &&
      gdk_pixbuf_save_to_buffer(pixbuf, &buf, &len, "png", nullptr, nullptr)) {
    fl_value_set_string_take(
        out, "png",
        fl_value_new_uint8_list(reinterpret_cast<const uint8_t*>(buf), len));
  }
  g_free(buf);
  clipboard_respond(call, out);
}

static void clipboard_got_png(GtkClipboard* cb, GtkSelectionData* d,
                              gpointer call) {
  gint len = 0;
  const guchar* data =
      d == nullptr ? nullptr : gtk_selection_data_get_data_with_length(d, &len);
  if (data == nullptr || len <= 0) {
    gtk_clipboard_request_image(cb, clipboard_got_image, call);
    return;
  }
  g_autoptr(FlValue) out = fl_value_new_map();
  fl_value_set_string_take(out, "png", fl_value_new_uint8_list(data, len));
  clipboard_respond(call, out);
}

static void clipboard_got_uris(GtkClipboard* cb, gchar** uris, gpointer call) {
  g_autoptr(FlValue) out = fl_value_new_map();
  if (uris != nullptr) {
    FlValue* list = fl_value_new_list();
    for (gchar** u = uris; *u != nullptr; u++) {
      fl_value_append_take(list, fl_value_new_string(*u));
    }
    fl_value_set_string_take(out, "uris", list);
  }
  clipboard_respond(call, out);
}

static void clipboard_got_targets(GtkClipboard* cb, GdkAtom* targets, gint n,
                                  gpointer call) {
  if (targets == nullptr || n <= 0) {
    clipboard_empty(call);
    return;
  }
  if (gtk_targets_include_uri(targets, n)) {
    gtk_clipboard_request_uris(cb, clipboard_got_uris, call);
    return;
  }
  if (!gtk_targets_include_image(targets, n, FALSE) ||
      gtk_targets_include_text(targets, n)) {
    clipboard_empty(call);
    return;
  }
  GdkAtom png = gdk_atom_intern_static_string("image/png");
  for (gint i = 0; i < n; i++) {
    if (targets[i] == png) {
      gtk_clipboard_request_contents(cb, png, clipboard_got_png, call);
      return;
    }
  }
  gtk_clipboard_request_image(cb, clipboard_got_image, call);
}

static void files_call_cb(FlMethodChannel* channel, FlMethodCall* call,
                          gpointer user_data) {
  if (g_strcmp0(fl_method_call_get_name(call), "clipboard") != 0) {
    fl_method_call_respond_not_implemented(call, nullptr);
    return;
  }
  gtk_clipboard_request_targets(gtk_clipboard_get(GDK_SELECTION_CLIPBOARD),
                                clipboard_got_targets, g_object_ref(call));
}

// Implements GApplication::activate.
static void my_application_activate(GApplication* application) {
  MyApplication* self = MY_APPLICATION(application);
  GtkWindow* window =
      GTK_WINDOW(gtk_application_window_new(GTK_APPLICATION(application)));

  // No system title bar: the app draws its own (window_manager, see
  // lib/core/desktop/desktop_window.dart). The title stays for the taskbar.
  gtk_window_set_title(window, "MegaChat");
  gtk_window_set_decorated(window, FALSE);

  // Window icon (X11 taskbars). GNOME on Wayland takes the icon from the
  // .desktop entry the app writes on start (desktop_entry.dart).
  g_autofree gchar* exe = g_file_read_link("/proc/self/exe", nullptr);
  if (exe != nullptr) {
    g_autofree gchar* dir = g_path_get_dirname(exe);
    g_autofree gchar* icon = g_build_filename(
        dir, "data", "flutter_assets", "assets", "icon", "megachat.png",
        nullptr);
    gtk_window_set_icon_from_file(window, icon, nullptr);
  }

  gtk_window_set_default_size(window, 1440, 900);

  g_autoptr(FlDartProject) project = fl_dart_project_new();
  fl_dart_project_set_dart_entrypoint_arguments(
      project, self->dart_entrypoint_arguments);

  FlView* view = fl_view_new(project);
  GdkRGBA background_color;
  // Background defaults to black, override it here if necessary, e.g. #00000000
  // for transparent.
  gdk_rgba_parse(&background_color, "#000000");
  fl_view_set_background_color(view, &background_color);
  gtk_widget_show(GTK_WIDGET(view));
  gtk_container_add(GTK_CONTAINER(window), GTK_WIDGET(view));

  // Show the window when Flutter renders.
  // Requires the view to be realized so we can start rendering.
  g_signal_connect_swapped(view, "first-frame", G_CALLBACK(first_frame_cb),
                           self);
  gtk_widget_realize(GTK_WIDGET(view));

  fl_register_plugins(FL_PLUGIN_REGISTRY(view));

  g_autoptr(FlStandardMethodCodec) codec = fl_standard_method_codec_new();
  g_clear_object(&self->files);
  self->files = fl_method_channel_new(
      fl_engine_get_binary_messenger(fl_view_get_engine(view)),
      "dev.megachat/files", FL_METHOD_CODEC(codec));
  fl_method_channel_set_method_call_handler(self->files, files_call_cb, self,
                                            nullptr);

  gtk_widget_grab_focus(GTK_WIDGET(view));
}

// Implements GApplication::local_command_line.
static gboolean my_application_local_command_line(GApplication* application,
                                                  gchar*** arguments,
                                                  int* exit_status) {
  MyApplication* self = MY_APPLICATION(application);
  // Strip out the first argument as it is the binary name.
  self->dart_entrypoint_arguments = g_strdupv(*arguments + 1);

  g_autoptr(GError) error = nullptr;
  if (!g_application_register(application, nullptr, &error)) {
    g_warning("Failed to register: %s", error->message);
    *exit_status = 1;
    return TRUE;
  }

  g_application_activate(application);
  *exit_status = 0;

  return TRUE;
}

// Implements GApplication::startup.
static void my_application_startup(GApplication* application) {
  // MyApplication* self = MY_APPLICATION(object);

  // Perform any actions required at application startup.

  G_APPLICATION_CLASS(my_application_parent_class)->startup(application);
}

// Implements GApplication::shutdown.
static void my_application_shutdown(GApplication* application) {
  // MyApplication* self = MY_APPLICATION(object);

  // Perform any actions required at application shutdown.

  G_APPLICATION_CLASS(my_application_parent_class)->shutdown(application);
}

// Implements GObject::dispose.
static void my_application_dispose(GObject* object) {
  MyApplication* self = MY_APPLICATION(object);
  g_clear_pointer(&self->dart_entrypoint_arguments, g_strfreev);
  g_clear_object(&self->files);
  G_OBJECT_CLASS(my_application_parent_class)->dispose(object);
}

static void my_application_class_init(MyApplicationClass* klass) {
  G_APPLICATION_CLASS(klass)->activate = my_application_activate;
  G_APPLICATION_CLASS(klass)->local_command_line =
      my_application_local_command_line;
  G_APPLICATION_CLASS(klass)->startup = my_application_startup;
  G_APPLICATION_CLASS(klass)->shutdown = my_application_shutdown;
  G_OBJECT_CLASS(klass)->dispose = my_application_dispose;
}

static void my_application_init(MyApplication* self) {}

MyApplication* my_application_new() {
  // Set the program name to the application ID, which helps various systems
  // like GTK and desktop environments map this running application to its
  // corresponding .desktop file. This ensures better integration by allowing
  // the application to be recognized beyond its binary name.
  g_set_prgname(APPLICATION_ID);

  return MY_APPLICATION(g_object_new(my_application_get_type(),
                                     "application-id", APPLICATION_ID, "flags",
                                     G_APPLICATION_NON_UNIQUE, nullptr));
}
