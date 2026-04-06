using System.Net.Sockets;
using System.Text;
using System.Text.Json;

namespace WizController.Services;

public record LightState(bool IsOn, int Brightness);

public class WizService
{
    private const int WizPort = 38899;

    public async Task<LightState> GetLightStateAsync(string ip)
    {
        var payload = new { method = "getPilot", @params = new { } };
        string jsonPayload = JsonSerializer.Serialize(payload);
        byte[] data = Encoding.UTF8.GetBytes(jsonPayload);

        using var udpClient = new UdpClient();
        using var cts = new CancellationTokenSource(TimeSpan.FromMilliseconds(2000));
        try
        {
            await udpClient.SendAsync(data, data.Length, ip, WizPort);
            var result = await udpClient.ReceiveAsync(cts.Token);
            var doc = JsonDocument.Parse(Encoding.UTF8.GetString(result.Buffer));
            var r = doc.RootElement.GetProperty("result");
            bool isOn = r.GetProperty("state").GetBoolean();
            int brightness = isOn && r.TryGetProperty("dimming", out var d) ? d.GetInt32() : 0;
            return new LightState(isOn, brightness);
        }
        catch
        {
            return new LightState(false, 0);
        }
    }

    public async Task SetLightStateAsync(string ip, bool state)
    {
        var payload = new
        {
            method = "setState",
            @params = new { state = state }
        };

        string jsonPayload = JsonSerializer.Serialize(payload);
        byte[] data = Encoding.UTF8.GetBytes(jsonPayload);

        using var udpClient = new UdpClient();
        try
        {
            await udpClient.SendAsync(data, data.Length, ip, WizPort);
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error sending to {ip}: {ex.Message}");
        }
    }

    public async Task SetBrightnessAsync(string ip, int brightness)
    {
        // Brightness range: 1-100 (0 = off, handled by SetLightStateAsync)
        var payload = new
        {
            method = "setPilot",
            @params = new { state = true, dimming = brightness }
        };

        string jsonPayload = JsonSerializer.Serialize(payload);
        byte[] data = Encoding.UTF8.GetBytes(jsonPayload);

        using var udpClient = new UdpClient();
        try
        {
            await udpClient.SendAsync(data, data.Length, ip, WizPort);
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error sending to {ip}: {ex.Message}");
        }
    }
}
