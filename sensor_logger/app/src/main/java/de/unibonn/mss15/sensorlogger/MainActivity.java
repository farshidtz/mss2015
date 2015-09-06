package de.unibonn.mss15.sensorlogger;

import android.app.Activity;
import android.app.ProgressDialog;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.os.AsyncTask;
import android.os.Bundle;
import android.os.IBinder;
import android.text.TextUtils;
import android.util.Log;
import android.view.View;
import android.view.inputmethod.InputMethodManager;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.EditText;
import android.widget.ScrollView;
import android.widget.Spinner;
import android.widget.Switch;
import android.widget.TextView;
import android.widget.Toast;
import android.widget.ToggleButton;
import org.apache.http.HttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.DefaultHttpClient;

import java.util.List;
import java.util.Vector;

public class MainActivity extends Activity {
    private final String SERVER_ADDR = "http://46.101.133.187:8529/sensors-data-collector/save";
    private final int PROC_SLICE = 5000; // number of entries in each marshal+POST request

    // UI objects
    private TextView logTxt;
    private ScrollView logScroll;
    private Spinner positionSpinner;
    private ProgressDialog pDialog;
    private TextView syncPendingTxt;
    private Button syncBtn;
    private EditText samplingPeriodTxt;
    private Switch predictionSwitch;
    private TextView positionLbl;
    private TextView syncStateLbl;

    // Service objects
    SensorLoggerService mService;
    boolean mBound = false;
    boolean predictionMode = false;

    // Storage
    private Storage storage;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        // UI objects
        logTxt = (TextView) findViewById(R.id.logTxt);
        logScroll = (ScrollView) findViewById(R.id.logScroll);
        syncPendingTxt = (TextView) findViewById(R.id.syncPendingTxt);
        syncBtn = (Button) findViewById(R.id.syncBtn);
        samplingPeriodTxt= (EditText) findViewById(R.id.samplingPeriodTxt);
        predictionSwitch = (Switch)  findViewById(R.id.predictionSwitch);
        positionLbl = (TextView) findViewById(R.id.positionLbl);
        syncStateLbl = (TextView) findViewById(R.id.syncStateLbl);
        // Spinner !
        positionSpinner = (Spinner) findViewById(R.id.positionSpinner);
        ArrayAdapter<CharSequence> adapter = ArrayAdapter.createFromResource(this, R.array.logging_contexts, android.R.layout.simple_spinner_item);
        adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        positionSpinner.setAdapter(adapter);

        storage = new Storage();
    }

/*    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        // Inflate the menu; this adds items to the action bar if it is present.
        getMenuInflater().inflate(R.menu.menu_main, menu);
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        // Handle action bar item clicks here. The action bar will
        // automatically handle clicks on the Home/Up button, so long
        // as you specify a parent activity in AndroidManifest.xml.
        int id = item.getItemId();

        //noinspection SimplifiableIfStatement
        if (id == R.id.action_settings) {
            return true;
        }

        return super.onOptionsItemSelected(item);
    }*/

    public void log(CharSequence text){
        logTxt.append(text + "\n");
        // Scroll to bottom
        logScroll.post(new Runnable() {
            public void run() {
                logScroll.smoothScrollTo(0, logScroll.getBottom());
            }
        });
    }

    public void startLoggerTglOnClick(View v) {
        boolean on = ((ToggleButton) v).isChecked();
        if (on) {
            positionSpinner.setEnabled(false);
            samplingPeriodTxt.setEnabled(false);
            syncBtn.setEnabled(false);
            predictionSwitch.setEnabled(false);

            // Start service
            startLoggerService();
            log("Service started.");
        } else {
            // Stop service
            stopLoggerService();

            log("Service stopped.");
            positionSpinner.setEnabled(true);
            samplingPeriodTxt.setEnabled(true);
            syncBtn.setEnabled(true);
            predictionSwitch.setEnabled(true);
        }
    }

    public void predictionSwitchOnClick(View v) {
        predictionMode = ((Switch) v).isChecked();
        if (predictionMode) {
            positionSpinner.setEnabled(false);
            syncBtn.setEnabled(false);
            positionLbl.setEnabled(false);
            syncStateLbl.setEnabled(false);
            syncPendingTxt.setEnabled(false);
            log("Switched to prediction mode.");
        } else {
            log("Switched to logging mode.");
            positionSpinner.setEnabled(true);
            syncBtn.setEnabled(true);
            positionLbl.setEnabled(true);
            syncStateLbl.setEnabled(true);
            syncPendingTxt.setEnabled(true);
        }
    }

    public void syncBtnOnClick(View v) {
        if(mBound){
            Toast.makeText(this, "Logging must be turned off before sync.", Toast.LENGTH_SHORT).show();
            return;
        }

        // Convert to JSON and call POST thread
        new PrepareAndPost().execute(storage);
    }

    public void startLoggerService(){
        // Bind to service
        Intent intent = new Intent(this, SensorLoggerService.class);
        intent.putExtra("SamplingPeriod", Integer.parseInt(samplingPeriodTxt.getText().toString()));
        intent.putExtra("PredictionMode", predictionMode);
        bindService(intent, mConnection, Context.BIND_AUTO_CREATE);
    }

    public void stopLoggerService(){
        // Unbind from the service
        if (mBound) {
            Storage newData = mService.stopLogging();
            storage.Append(newData);
            unbindService(mConnection);
            mBound = false;
            syncPendingTxt.setText(Integer.toString(storage.Size()));
            log("Retrieved " + Integer.toString(newData.Size()) + " entries.");
        }
    }

    /** Defines callbacks for service binding, passed to bindService() */
    private ServiceConnection mConnection = new ServiceConnection() {

        @Override
        public void onServiceConnected(ComponentName className, IBinder service) {
            // We've bound to LocalService, cast the IBinder and get LocalService instance
            SensorLoggerService.LocalBinder binder = (SensorLoggerService.LocalBinder) service;
            mService = binder.getService();
            mBound = true;

            // Get spinner choice and start logging
            String chosenContext = positionSpinner.getSelectedItem().toString();
            mService.startLogging(chosenContext.replace(" ", ""));
        }

        @Override
        public void onServiceDisconnected(ComponentName arg0) { mBound = false; }
    };

    // Marshalls Storage to JSON string and calls POST
    class PrepareAndPost extends AsyncTask<Storage, Integer, String> {
        private int progressMax;

        @Override
        protected void onPreExecute() {
            super.onPreExecute();
            log("Posting " + storage.Size() + " entries in " + PROC_SLICE + " blocks ...");
            // Showing progress dialog
            pDialog = new ProgressDialog(MainActivity.this);
            pDialog.setMessage("Please wait ...");
            pDialog.setCancelable(false);
            //pDialog.show();
            pDialog.setProgressStyle(ProgressDialog.STYLE_HORIZONTAL);
            pDialog.setProgress(0);
            progressMax = storage.Size();
            pDialog.setMax(progressMax);
            pDialog.show();
        }

        @Override
        protected String doInBackground(Storage... s) {
            List<String> messages = new Vector<String>();
            while(s[0].Size() != 0) {
                publishProgress(s[0].Size());

                // Convert a slice of storage to json
                String json = s[0].PopFrontJSON(PROC_SLICE);

                //String responseStatusLine = "";
                try {
                    DefaultHttpClient httpclient = new DefaultHttpClient();
                    HttpPost httppostreq = new HttpPost(SERVER_ADDR);
                    StringEntity se = new StringEntity(json);
                    httppostreq.setEntity(se);
                    HttpResponse httpresponse = httpclient.execute(httppostreq);
                    //responseStatusLine = httpresponse.getStatusLine().toString();
                    if (httpresponse.getStatusLine().getStatusCode() == 200) {
                        messages.add(httpresponse.getStatusLine().toString());
                    }
                    else {
                        throw new Exception(httpresponse.getStatusLine().toString());
                    }
                } catch (Exception e) {
                    storage.Recover();
                    messages.add(e.getMessage());
                    return TextUtils.join("\n", messages);
                }
            }
            return TextUtils.join("\n", messages);
        }

        @Override
        protected void onProgressUpdate(Integer... progress) {
            pDialog.setProgress(progressMax-progress[0]);
        }

        @Override
        protected void onPostExecute(String message) {
            super.onPostExecute(message);
            Log.d("JSON", message);
            log(message);

            // Dismiss the progress dialog
            if (pDialog.isShowing())
                pDialog.dismiss();

            syncPendingTxt.setText(Integer.toString(storage.Size()));
            log("Pending entries: " + Integer.toString(storage.Size()));
        }
    }
}