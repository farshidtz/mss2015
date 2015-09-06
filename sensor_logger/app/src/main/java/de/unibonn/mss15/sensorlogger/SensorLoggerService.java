package de.unibonn.mss15.sensorlogger;

import android.app.Service;
import android.content.Intent;
import android.hardware.Sensor;
import android.hardware.SensorEvent;
import android.hardware.SensorEventListener;
import android.hardware.SensorManager;
import android.os.AsyncTask;
import android.os.Binder;
import android.os.IBinder;
import android.util.Log;
import android.widget.Toast;
import android.os.PowerManager;


import org.eclipse.paho.android.service.MqttAndroidClient;
import org.eclipse.paho.client.mqttv3.IMqttActionListener;
import org.eclipse.paho.client.mqttv3.IMqttDeliveryToken;
import org.eclipse.paho.client.mqttv3.IMqttToken;
import org.eclipse.paho.client.mqttv3.MqttCallback;
import org.eclipse.paho.client.mqttv3.MqttClient;
import org.eclipse.paho.client.mqttv3.MqttConnectOptions;
import org.eclipse.paho.client.mqttv3.MqttException;
import org.eclipse.paho.client.mqttv3.MqttMessage;
import org.eclipse.paho.client.mqttv3.MqttPersistenceException;
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence;

import java.util.TimeZone;

/**
 * Service to collect sensor data and return on stop
 */

public class SensorLoggerService extends Service {
    // Binder given to client
    private final IBinder mBinder = new LocalBinder();

    // System attributes
    private SensorManager sensorManager = null;
    private SensorEventListener sensorListener;

    // Sensor storage
    private Storage storage = new Storage();
    private int samplingRate; // micro seconds

    // Power management
    PowerManager powerManager;
    PowerManager.WakeLock wakeLock;

    // Local attributes
    private long serviceStartTime;
    private long serviceTime;
    private int timezoneOffset = TimeZone.getDefault().getRawOffset() + TimeZone.getDefault().getDSTSavings();

    // MQTT Client
    MQTTClient mqtt;

    /**
     * Class used for the client Binder.  Because we know this service always
     * runs in the same process as its clients, we don't need to deal with IPC.
     */
    public class LocalBinder extends Binder {
        SensorLoggerService getService() {
            // Return this instance of LocalService so clients can call public methods
            return SensorLoggerService.this;
        }
    }

    @Override
    public IBinder onBind(Intent intent) {
        // Get sampling rate from Activity
        samplingRate = intent.getIntExtra("SamplingRate", 500)*1000;

        // Setup power manager
        powerManager = (PowerManager) getSystemService(POWER_SERVICE);
        wakeLock = powerManager.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "WAKE_LOCK");

        // MQTT client
        mqtt = new MQTTClient();

        sensorManager = (SensorManager) getSystemService(SENSOR_SERVICE);
        sensorListener = new SensorEventListener() {
            @Override
            public void onAccuracyChanged(Sensor arg0, int arg1) {}

            @Override
            public void onSensorChanged(SensorEvent event) {
                new SensorEventLoggerTask().execute(event);
            }
        };

        return mBinder;
    }

    @Override
    public void onDestroy () {
        // stop the sensor and service
        Toast.makeText(this, "Logger Service Stopped.", Toast.LENGTH_SHORT).show();
        Log.v("Logger", "Byebye!");
        stopSelf();
    }


    // Record data
    private class SensorEventLoggerTask extends AsyncTask<SensorEvent, Void, Void> {
        @Override
        protected Void doInBackground(SensorEvent... events) {
            SensorEvent event = events[0];
            long t = System.currentTimeMillis() + timezoneOffset;

            // Error rate
            serviceTime = (int) ((t - serviceStartTime)/1000);
            int e = 0;
            if(serviceTime <= 10) {
                e = 10 - (int) serviceTime;
                Log.d("Error rate:", Integer.toString(e));
            }



            Sensor sensor = event.sensor;
            if (sensor.getType() == Sensor.TYPE_ACCELEROMETER) {
                float x = event.values[0];
                float y = event.values[1];
                float z = event.values[2];
                Log.d("Acc:", Float.toString(x) + "," + Float.toString(y) + "," + Float.toString(z));
                storage.AddEntry(t, e, "acc", 3, event.values);
                mqtt.publicJSON(t, e, "acc", 3, event.values);

            } else if (sensor.getType() == Sensor.TYPE_GYROSCOPE) {
                float x = event.values[0];
                float y = event.values[1];
                float z = event.values[2];
                Log.d("Gyro:", Float.toString(x) + "," + Float.toString(y) + "," + Float.toString(z));
                storage.AddEntry(t, e, "gyro", 3, event.values);
                mqtt.publicJSON(t, e, "gyro", 3, event.values);

            } else if (sensor.getType() == Sensor.TYPE_PRESSURE) {
                float v = event.values[0];
                Log.d("Pressure", Float.toString(v));
                storage.AddEntry(t, e, "pressure", 1,event.values);
                mqtt.publicJSON(t, e, "pressure", 1,event.values);

            } else if (sensor.getType() == Sensor.TYPE_LIGHT) {
                float v = event.values[0];
                Log.d("Light", Float.toString(v));
                storage.AddEntry(t, e, "light", 1, event.values);
                mqtt.publicJSON(t, e, "light", 1, event.values);

            } else if (sensor.getType() == Sensor.TYPE_PROXIMITY) {
                float v = event.values[0];
                Log.d("Proximity", Float.toString(v));
                storage.AddEntry(t, e, "proximity", 1,event.values);
                mqtt.publicJSON(t, e, "proximity", 1,event.values);

            } else if (sensor.getType() == Sensor.TYPE_LINEAR_ACCELERATION) {
                float x = event.values[0];
                float y = event.values[1];
                float z = event.values[2];
                Log.d("LinAcc:", Float.toString(x) + "," + Float.toString(y) + "," + Float.toString(z));
                storage.AddEntry(t, e, "linacc", 3, event.values);
                mqtt.publicJSON(t, e, "linacc", 3, event.values);

            } else if (sensor.getType() == Sensor.TYPE_ROTATION_VECTOR) {
                float x = event.values[0];
                float y = event.values[1];
                float z = event.values[2];
                //float zz = event.values[3];
                Log.d("Rotation:", Float.toString(x) + "," + Float.toString(y) + "," + Float.toString(z));
                storage.AddEntry(t, e, "rotation", 3, event.values);
                mqtt.publicJSON(t, e, "rotation", 3, event.values);

            } else if (sensor.getType() == Sensor.TYPE_GRAVITY) {
                float x = event.values[0];
                float y = event.values[1];
                float z = event.values[2];
                //float zz = event.values[3];
                Log.d("Gravity:", Float.toString(x) + "," + Float.toString(y) + "," + Float.toString(z));
                storage.AddEntry(t, e, "gravity", 3, event.values);
                mqtt.publicJSON(t, e, "gravity", 3, event.values);
            }

            return null;
        }
    }

    /********************** Methods for the client **********************/

    // Register to sensors
    public void startLogging(String phonePos) {
        // Store the spinner's value (phone position)
        storage.SetPosition(phonePos);

        // Keep CPU awake
        wakeLock.acquire();

        Toast.makeText(this, "Sampling every "+ Integer.toString(samplingRate/1000)+"ms", Toast.LENGTH_SHORT).show();
        Log.v("Sampling period", Integer.toString(samplingRate / 1000) + "ms");

        // Record service start time
        serviceStartTime = System.currentTimeMillis()+timezoneOffset;
        // A header above all entries at each logging session
        storage.AddEntry(serviceStartTime, 100, "system", 0);

        // Register sensor listeners
        //sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER), samplingRate, samplingRate);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_GYROSCOPE), samplingRate, samplingRate);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_PRESSURE), samplingRate, samplingRate);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_LIGHT), samplingRate, samplingRate);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_PROXIMITY), samplingRate, samplingRate);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_LINEAR_ACCELERATION), samplingRate, samplingRate);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_ROTATION_VECTOR ), samplingRate, samplingRate);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_GRAVITY ), samplingRate, samplingRate);
    }

    // Unregister from sensors and return data
    public Storage stopLogging(){
        // Unregister sensor listeners
        sensorManager.unregisterListener(sensorListener);

        // Release CPU Lock
        wakeLock.release();

        // Reset error rates
        storage.ResetErrorRates(System.currentTimeMillis()+timezoneOffset);
        return storage;
    }


}