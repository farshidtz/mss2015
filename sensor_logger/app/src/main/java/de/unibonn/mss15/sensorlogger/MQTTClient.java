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
 * Connect to and publish json messages to an MQTT broker
 */
public class MQTTClient extends SensorLoggerService implements MqttCallback {

    Context context;
    MemoryPersistence memPer;
    MqttClient client;
    final String BrokerURI = "tcp://192.168.1.42:1883";
    //final String BrokerURI = "tcp://iot.eclipse.org:1883";

    public MQTTClient(Context context){
        this.context = context;
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
            e.printStackTrace();
            Toast.makeText(context, "MQTT Error: "+ e.getMessage(), Toast.LENGTH_SHORT).show();
        }

        MqttConnectOptions options = new MqttConnectOptions();
        try
        {
            client.connect(options);
        }
        catch (MqttException e)
        {
            Log.d(getClass().getCanonicalName(), "Connection attempt failed with reason code = " + e.getReasonCode() + ":" + e.getCause());
            Toast.makeText(context, "MQTT Error: Connection attempt failed: "+ e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }

    public void disconnect(){
        try {
            client.disconnect();
        } catch (MqttException e) {
            e.printStackTrace();
            Toast.makeText(context, "MQTT Error: "+ e.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }

    public void publish(String msg){
        try
        {
            MqttMessage message = new MqttMessage();
            message.setQos(1);
            message.setPayload(msg.getBytes());
            client.publish("sensors/data", message);
        }
        catch (MqttException e)
        {
            Log.d(getClass().getCanonicalName(), "Publish failed with reason code = " + e.getReasonCode());
            Toast.makeText(context, "MQTT Error: Publish failed: "+ e.getMessage(), Toast.LENGTH_SHORT).show();
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
            Toast.makeText(context, "MQTT Error: Json marshaller failed: "+ ex.getMessage(), Toast.LENGTH_SHORT).show();
        }
    }

    @Override
    public void connectionLost(Throwable cause)
    {
        Log.d("MQTT", "MQTT Server connection lost" + cause.getMessage());
        Toast.makeText(context, "MQTT Server connection lost: "+ cause.getMessage(), Toast.LENGTH_SHORT).show();
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
