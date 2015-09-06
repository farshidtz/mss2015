package de.unibonn.mss15.sensorlogger;

import android.content.Context;
import android.util.Log;
import android.widget.Toast;


import org.eclipse.paho.client.mqttv3.IMqttDeliveryToken;
import org.eclipse.paho.client.mqttv3.MqttCallback;
import org.eclipse.paho.client.mqttv3.MqttClient;
import org.eclipse.paho.client.mqttv3.MqttConnectOptions;
import org.eclipse.paho.client.mqttv3.MqttException;
import org.eclipse.paho.client.mqttv3.MqttMessage;
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence;
import com.google.gson.Gson;

/**
 * Connect and publish json messages to an MQTT broker
 */
public class MQTTClient implements MqttCallback {

    Context context;
    MemoryPersistence memPer;
    MqttClient client;
    String BrokerURI;
    Boolean connected = true;

    public MQTTClient(Context context, String BrokerURI){
        this.context = context;
        this.BrokerURI = BrokerURI;
    }

    public void connect(){
        memPer = new MemoryPersistence();
        try
        {
            client = new MqttClient(BrokerURI, MqttClient.generateClientId(), null);
            client.setCallback(this);
        }
        catch (MqttException e)
        {
            Log.d(getClass().getCanonicalName(), "Failed to create MqttClient: " + e.getMessage());
            Toast.makeText(context, "MQTT Error: "+ e.getMessage(), Toast.LENGTH_LONG).show();
            connected = false;
        }

        MqttConnectOptions options = new MqttConnectOptions();
        try
        {
            client.setTimeToWait(3000);
            client.connect(options);
        }
        catch (MqttException|NumberFormatException e)
        {
            Log.d(getClass().getCanonicalName(), "Connection attempt failed: :" + e.getMessage());
            Toast.makeText(context, "MQTT Error: Connection attempt failed: "+ e.getMessage(), Toast.LENGTH_LONG).show();
            connected = false;
        }
    }

    public void disconnect(){
        if (!connected)
            return;
        try {
            client.disconnect();
        } catch (MqttException e) {
            e.printStackTrace();
            Toast.makeText(context, "MQTT Error: "+ e.getMessage(), Toast.LENGTH_LONG).show();
        }
    }

    public void publish(String msg){
        if (!connected)
            return;
        try
        {
            MqttMessage message = new MqttMessage();
            message.setQos(1);
            message.setPayload(msg.getBytes());
            client.publish("mss2015/sensors/data", message);
        }
        catch (MqttException e)
        {
            Log.d(getClass().getCanonicalName(), "Publish failed with reason code = " + e.getReasonCode());
            //Toast.makeText(context, "MQTT Error: Publish failed: "+ e.getMessage(), Toast.LENGTH_LONG).show();
        }
    }

    public void publishJSON(long t, int e, String n, int axes, float... values){
        try{
            Entry entry = new Entry("?",t,e,n,axes,values);
            Gson gson = new Gson();
            String json = gson.toJson(entry);
            publish(json);
        }
        catch(IllegalArgumentException|SecurityException|IllegalAccessException|NoSuchFieldException ex){
            Log.d("Exception", ex.getMessage());
            Toast.makeText(context, "MQTT Error: Json marshaller failed: "+ ex.getMessage(), Toast.LENGTH_LONG).show();
        }
    }

    @Override
    public void connectionLost(Throwable cause)
    {
        Log.d("MQTT", "MQTT Server connection lost" + cause.getMessage());
        Toast.makeText(context, "MQTT Server connection lost: "+ cause.getMessage(), Toast.LENGTH_LONG).show();
    }
    @Override
    public void messageArrived(String topic, MqttMessage message)
    {
        Log.d("MQTT", "Message arrived:" + topic + ":" + message.toString());
    }
    @Override
    public void deliveryComplete(IMqttDeliveryToken token)
    {
        Log.d("MQTT", "Delivery complete");
    }
}
