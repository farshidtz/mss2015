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

import java.util.TimeZone;

/**
 * Service to collect sensor data
 * Returns on logging mode
 * Publishes to MQTT on prediction mode
 */

public class SensorLoggerService extends Service {
    // Binder given to client
    private final IBinder mBinder = new LocalBinder();

    // System attributes
    private SensorManager sensorManager = null;
    private SensorEventListener sensorListener;

    // Sensor storage
    private Storage storage = new Storage();
    private int samplingPeriod; // micro seconds

    // Power management
    PowerManager powerManager;
    PowerManager.WakeLock wakeLock;

    // Local attributes
    private long serviceStartTime;
    private long serviceTime;
    private  boolean predictionMode;
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

    // Prepare the service
    @Override
    public IBinder onBind(Intent intent) {
        // Get data from Activity
        samplingPeriod = intent.getIntExtra("SamplingPeriod", 500)*1000;
        predictionMode = intent.getBooleanExtra("PredictionMode", false);

        // Setup mqtt client with the service context
        mqtt = new MQTTClient(this);

        // Setup power manager
        powerManager = (PowerManager) getSystemService(POWER_SERVICE);
        wakeLock = powerManager.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "WAKE_LOCK");

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
        // destroy the service
        Toast.makeText(this, "Logger Service Stopped.", Toast.LENGTH_SHORT).show();
        Log.d("Logger", "Byebye!");
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


            // Record sensor values
            Sensor sensor = event.sensor;
            if (sensor.getType() == Sensor.TYPE_ACCELEROMETER) {
                Log.d("Acc:", Float.toString(event.values[0]) + "," + Float.toString(event.values[1]) + "," + Float.toString(event.values[2]));
                if(predictionMode){
                    mqtt.publishJSON(t, e, "acc", 3, event.values);
                } else {
                    storage.AddEntry(t, e, "acc", 3, event.values);
                }

            } else if (sensor.getType() == Sensor.TYPE_GYROSCOPE) {
                Log.d("Gyro:", Float.toString(event.values[0]) + "," + Float.toString(event.values[1]) + "," + Float.toString(event.values[2]));
                if(predictionMode) {
                    mqtt.publishJSON(t, e, "gyro", 3, event.values);
                } else {
                    storage.AddEntry(t, e, "gyro", 3, event.values);
                }

            } else if (sensor.getType() == Sensor.TYPE_PRESSURE) {
                Log.d("Pressure", Float.toString(event.values[0]));
                if(predictionMode) {
                    mqtt.publishJSON(t, e, "pressure", 1, event.values);
                } else {
                    storage.AddEntry(t, e, "pressure", 1,event.values);
                }

            } else if (sensor.getType() == Sensor.TYPE_LIGHT) {
                Log.d("Light", Float.toString(event.values[0]));
                if(predictionMode) {
                    mqtt.publishJSON(t, e, "light", 1, event.values);
                } else {
                    storage.AddEntry(t, e, "light", 1, event.values);
                }

            } else if (sensor.getType() == Sensor.TYPE_PROXIMITY) {
                Log.d("Proximity", Float.toString(event.values[0]));
                if(predictionMode) {
                    mqtt.publishJSON(t, e, "proximity", 1, event.values);
                } else {
                    storage.AddEntry(t, e, "proximity", 1,event.values);
                }

            } else if (sensor.getType() == Sensor.TYPE_LINEAR_ACCELERATION) {
                Log.d("LinAcc:", Float.toString(event.values[0]) + "," + Float.toString(event.values[1]) + "," + Float.toString(event.values[2]));
                if(predictionMode) {
                    mqtt.publishJSON(t, e, "linacc", 3, event.values);
                } else {
                    storage.AddEntry(t, e, "linacc", 3, event.values);
                }

            } else if (sensor.getType() == Sensor.TYPE_ROTATION_VECTOR) {
                Log.d("Rotation:", Float.toString(event.values[0]) + "," + Float.toString(event.values[1]) + "," + Float.toString(event.values[2]));
                if(predictionMode) {
                    mqtt.publishJSON(t, e, "rotation", 3, event.values);
                } else {
                    storage.AddEntry(t, e, "rotation", 3, event.values);
                }

            } else if (sensor.getType() == Sensor.TYPE_GRAVITY) {
                Log.d("Gravity:", Float.toString(event.values[0]) + "," + Float.toString(event.values[1]) + "," + Float.toString(event.values[2]));
                if(predictionMode) {
                    mqtt.publishJSON(t, e, "gravity", 3, event.values);
                } else {
                    storage.AddEntry(t, e, "gravity", 3, event.values);
                }

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

        // start MQTT client
        mqtt.connect();

        Toast.makeText(this, "Sampling every "+ Integer.toString(samplingPeriod/1000)+"ms", Toast.LENGTH_SHORT).show();
        Log.v("Sampling period", Integer.toString(samplingPeriod / 1000) + "ms");

        // Record service start time
        serviceStartTime = System.currentTimeMillis()+timezoneOffset;
        // A header above all entries at each session
        if(predictionMode){
            mqtt.publishJSON(serviceStartTime, 100, "system", 0);
        } else {
            storage.AddEntry(serviceStartTime, 100, "system", 0);
        }

        // Register sensor listeners
        //sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_GYROSCOPE), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_PRESSURE), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_LIGHT), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_PROXIMITY), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_LINEAR_ACCELERATION), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_ROTATION_VECTOR ), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_GRAVITY ), samplingPeriod, samplingPeriod);
    }

    // Unregister from sensors and return data
    public Storage stopLogging(){
        // Unregister sensor listeners
        sensorManager.unregisterListener(sensorListener);

        // Release CPU Lock
        wakeLock.release();

        // Stop mqtt client
        mqtt.disconnect();

        // Reset error rates
        storage.ResetErrorRates(System.currentTimeMillis()+timezoneOffset);
        return storage;
    }


}