package de.unibonn.mss15.sensorlogger;

import android.app.Activity;
import android.app.ProgressDialog;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.graphics.Color;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.os.AsyncTask;
import android.os.Bundle;
import android.os.IBinder;
import android.os.Messenger;
import android.text.method.ScrollingMovementMethod;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.ScrollView;
import android.widget.Spinner;
import android.widget.Switch;
import android.widget.TextView;
import android.widget.Toast;
import android.widget.ToggleButton;
import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.NameValuePair;
import org.apache.http.client.ClientProtocolException;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.DefaultHttpClient;
import org.apache.http.message.BasicNameValuePair;
import org.apache.http.protocol.HTTP;
import org.apache.http.util.EntityUtils;
import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.text.ParseException;
import java.util.Timer;


public class MainActivity extends Activity {
    // UI objects
    private TextView logTxt;
    private ScrollView logScroll;
    private Spinner spinner;
    private ProgressDialog pDialog;

    // Service objects
    SensorLoggerService mService;
    boolean mBound = false;

    // Storage
    private Storage storage;


    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        logTxt = (TextView) findViewById(R.id.logTxt);
        logScroll = (ScrollView) findViewById(R.id.logScroll);
        // Spinner !
        spinner = (Spinner) findViewById(R.id.logcontextSpinner);
        ArrayAdapter<CharSequence> adapter = ArrayAdapter.createFromResource(this, R.array.logging_contexts, android.R.layout.simple_spinner_item);
        adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        spinner.setAdapter(adapter);

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
            // Start service
            startLoggerService();
            log("Service started.");
        } else {
            // Stop service
            stopLoggerService();
            log("Service stopped.");
        }
    }

    public void syncBtnOnClick(View v) {
        if(mBound){
            Toast.makeText(this, "Logging must be turned off before sync.", Toast.LENGTH_SHORT).show();
            return;
        }
        // Post Entries
        String json = storage.ToJSON();
        //log("JSON" + json);
        Log.v("JSON", json);
        new PostData().execute(json);
    }

    public void startLoggerService(){
        // Bind to service
        Intent intent = new Intent(this, SensorLoggerService.class);
        bindService(intent, mConnection, Context.BIND_AUTO_CREATE);
    }

    public void stopLoggerService(){
        // Unbind from the service
        if (mBound) {
            storage = mService.stopLogging();
            unbindService(mConnection);
            mBound = false;
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
            String chosenContext = spinner.getSelectedItem().toString();
            mService.startLogging(chosenContext.replace(" ", ""));
        }

        @Override
        public void onServiceDisconnected(ComponentName arg0) {
            mBound = false;
        }
    };


    class PostData extends AsyncTask<String, String, String> {

        @Override
        protected void onPreExecute() {
            super.onPreExecute();
            // Showing progress dialog
            pDialog = new ProgressDialog(MainActivity.this);
            pDialog.setMessage("Please wait...");
            pDialog.setCancelable(false);
            pDialog.show();
        }

        //private Exception exception;
        private String responseStatusLine = "";

        protected String doInBackground(String... bodies) {
            try {
                DefaultHttpClient httpclient = new DefaultHttpClient();
                HttpPost httppostreq = new HttpPost("http://46.101.133.187:8529/sensors-data-collector/save");
                StringEntity se = new StringEntity(bodies[0]);
                httppostreq.setEntity(se);
                HttpResponse httpresponse = httpclient.execute(httppostreq);
                responseStatusLine = httpresponse.getStatusLine().toString();
            } catch (Exception e) {
                return e.getMessage();
            }
            return responseStatusLine;
        }

        protected void onPostExecute(String response) {
            super.onPostExecute(response);
            final String res = response;

            // Dismiss the progress dialog
            if (pDialog.isShowing())
                pDialog.dismiss();

            Log.v("HTTP", res);
            log(res);
            runOnUiThread(new Runnable() {
                public void run() {
                    Toast.makeText(getApplicationContext(), res, Toast.LENGTH_SHORT).show();
                }
            });
        }
    }
}
