/*
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

package dev.pgnd.cove;

import android.os.Bundle;
import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {
    @Override
    public void onCreate(Bundle savedInstanceState) {
        registerPlugin(HealthConnectPlugin.class);
        super.onCreate(savedInstanceState);
    }
}
