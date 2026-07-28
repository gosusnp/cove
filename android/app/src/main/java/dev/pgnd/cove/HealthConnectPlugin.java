/*
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

package dev.pgnd.cove;

import android.content.Intent;
import android.content.pm.PackageManager;
import androidx.activity.result.ActivityResult;
import androidx.health.connect.client.HealthConnectClient;
import androidx.health.connect.client.PermissionController;
import com.getcapacitor.JSObject;
import com.getcapacitor.Plugin;
import com.getcapacitor.PluginCall;
import com.getcapacitor.PluginMethod;
import com.getcapacitor.annotation.ActivityCallback;
import com.getcapacitor.annotation.CapacitorPlugin;
import java.util.Collections;

@CapacitorPlugin(name = "HealthConnect")
public class HealthConnectPlugin extends Plugin {

    // Permission string for writing exercise sessions to Health Connect.
    private static final String WRITE_EXERCISE = "android.permission.health.WRITE_EXERCISE";

    @PluginMethod
    public void isAvailable(PluginCall call) {
        int status = HealthConnectClient.getSdkStatus(
                getContext(), "com.google.android.apps.healthdata");
        JSObject result = new JSObject();
        result.put("available", status == HealthConnectClient.SDK_AVAILABLE);
        call.resolve(result);
    }

    @PluginMethod
    public void requestPermission(PluginCall call) {
        saveCall(call);
        java.util.Set<String> permissions = Collections.singleton(WRITE_EXERCISE);
        android.content.Intent intent = PermissionController
                .createRequestPermissionResultContract()
                .createIntent(getContext(), permissions);
        startActivityForResult(call, intent, "handlePermissionResult");
    }

    @ActivityCallback
    private void handlePermissionResult(PluginCall call, ActivityResult result) {
        if (call == null) return;
        boolean granted = getContext().checkSelfPermission(WRITE_EXERCISE)
                == PackageManager.PERMISSION_GRANTED;
        JSObject ret = new JSObject();
        ret.put("granted", granted);
        call.resolve(ret);
    }

    @PluginMethod
    public void openSettings(PluginCall call) {
        // Opens the Health Connect settings home where users can manage app permissions.
        Intent intent = new Intent("android.health.connect.action.HEALTH_HOME_SETTINGS");
        getActivity().startActivity(intent);
        call.resolve();
    }
}
