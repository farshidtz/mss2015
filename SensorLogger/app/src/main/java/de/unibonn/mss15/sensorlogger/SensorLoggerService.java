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

import java.util.Random;

public class SensorLoggerService extends Service {
    // Binder given to client
    private final IBinder mBinder = new LocalBinder();

    // Sensor system variables
    private SensorManager sensorManager = null;
    private SensorEventListener sensorListener;

    // Sensor storage
    private Storage storage = new Storage();
    private int samplingHz = 2;
    private int samplingPeriod = (1000/samplingHz)*1000; // micro seconds

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
            long t = event.timestamp;

            Sensor sensor = event.sensor;
            if (sensor.getType() == Sensor.TYPE_ACCELEROMETER) {
                float x = event.values[0];
                float y = event.values[1];
                float z = event.values[2];
                Log.v("Acc:", Float.toString(x) + "," + Float.toString(y) + "," + Float.toString(z));
                storage.AddEntry(t, "acc", 3, event.values);

            } else if (sensor.getType() == Sensor.TYPE_GYROSCOPE) {
                float x = event.values[0];
                float y = event.values[1];
                float z = event.values[2];
                Log.v("Gyro:", Float.toString(x) + "," + Float.toString(y) + "," + Float.toString(z));
                storage.AddEntry(t,"gyro",3,event.values);

            } else if (sensor.getType() == Sensor.TYPE_PRESSURE) {
                float v = event.values[0];
                Log.v("Baro", Float.toString(v));

                storage.AddEntry(t,"baro",1,event.values);
            } else if (sensor.getType() == Sensor.TYPE_LIGHT) {
                float v = event.values[0];
                Log.v("Light", Float.toString(v));
                storage.AddEntry(t,"light",1,event.values);

            } else if (sensor.getType() == Sensor.TYPE_PROXIMITY) {
                float v = event.values[0];
                Log.v("Proximity", Float.toString(v));
                storage.AddEntry(t,"proxi",1,event.values);

            }

            return null;
        }
    }

    /********************** Methods for the client **********************/

    // Register to sensors
    public void startLogging(String phonePos) {
        // Store the spinner's value
        storage.Position = phonePos;

        Toast.makeText(this, "Sampling every "+ Integer.toString(samplingPeriod/1000)+"ms", Toast.LENGTH_SHORT).show();
        Log.v("Sampling period", Integer.toString(samplingPeriod/1000)+"ms");

        // SensorManager.SENSOR_DELAY_GAME
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_GYROSCOPE), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_PRESSURE), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_LIGHT), samplingPeriod, samplingPeriod);
        sensorManager.registerListener(sensorListener, sensorManager.getDefaultSensor(Sensor.TYPE_PROXIMITY), samplingPeriod, samplingPeriod);
    }

    // Unregister from sensors and return data
    public Storage stopLogging(){
        sensorManager.unregisterListener(sensorListener);
        return storage;
    }


}




/*        new Thread(new Runnable() {
            public void run() {
                for (int i=0; i<100; i++) {
                    ToneGenerator toneG = new ToneGenerator(AudioManager.STREAM_ALARM, 100);
                    toneG.startTone(ToneGenerator.TONE_CDMA_ALERT_CALL_GUARD, 10);
                    try {
                        Thread.sleep(4000);
                    } catch (InterruptedException e) {
                        e.printStackTrace();
                    }
                }
            }
        }).start();*/