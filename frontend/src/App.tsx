import { AxiosError } from "axios"
import { axiosAPI } from "./utils/axiosAPI"
import { useRef, useState } from "react"
import toast from "react-hot-toast"

export default function App() {
    const [isLoaded, setIsLoaded] = useState(false)
    const [imageData, setImageData] = useState<string | null>(null)
    const [percentUpload, setPercentUpload] = useState(0)
    const [check45, setCheck45] = useState(false)
    const formRef = useRef<HTMLFormElement>(null)

    const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        const form = e.currentTarget as HTMLFormElement
        const file = form.elements.namedItem('file') as HTMLInputElement
        const rotateAngle = form.elements.namedItem('rotateAngle') as HTMLInputElement
        const scale = form.elements.namedItem('scale') as HTMLInputElement

        const data = new FormData()
        data.append('image', file.files![0])
        data.append('rotateAngle', rotateAngle.value)
        data.append('scale', scale.value)

        setIsLoaded(true)
        try {
            const urlApi = check45 ? '/api/45/generate' : '/api/916/generate'
            const res = await axiosAPI.post(urlApi, data, {
                responseType: 'blob',
                onUploadProgress: (e) => {
                    setPercentUpload(Math.round((e.loaded * 100) / (e.total ?? 1)))
                }
            })
            const url = URL.createObjectURL(res.data)
            setImageData(url)
            setIsLoaded(false)
        } catch (error) {
            if (error instanceof AxiosError) {
                // blob to text
                const reader = new FileReader()
                reader.onload = () => {
                    toast.error(JSON.parse(reader.result as string).error)
                    setIsLoaded(false)
                }
                reader.readAsText(error.response?.data)
            }
        }

    }

    const fixRotation = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault()
        const form = formRef.current as HTMLFormElement
        const rotateAngle = form.elements.namedItem('rotateAngle') as HTMLInputElement
        rotateAngle.value = '270'
    }

    return (
        <div className="min-h-screen flex flex-col items-center">
            <div className="h-full p-3 flex flex-col max-w-2xl w-full mt-16">
                <h1 className="text-3xl font-bold">
                    PolStory
                </h1>
                <form onSubmit={onSubmit} ref={formRef} className="flex flex-col w-full">
                    <input type="file" name="file" accept=".jpg, .jpeg, .png, .JPEG, .JPG, .PNG" className="text-zinc-400 cursor-pointer" />
                    <div className="flex flex-col w-full">
                        <div className="flex gap-2 mt-3">
                            <div className="flex flex-col w-full">
                                <div>Rotate Angle</div>
                                <input type="number" name="rotateAngle" className="border px-2 py-1 rounded-lg focus:outline-none w-full" placeholder="Rotate Angle" defaultValue={0} />
                            </div>
                            <div className="flex flex-col w-full">
                                <div>Scale</div>
                                <input type="number" name="scale" className="border px-2 py-1 rounded-lg focus:outline-none w-full" placeholder="Scale" defaultValue={1} />
                            </div>
                        </div>
                        <div className="flex justify-between mt-3 items-center">
                            <div className="flex justify-start gap-1" onClick={() => setCheck45(!check45)}>
                                <input type="checkbox" checked={check45} onChange={() => setCheck45(!check45)} />
                                <div>4:5</div>
                            </div>
                            <button className="bg-cyan-700 text-white px-2 py-1 rounded-lg hover:bg-cyan-800 duration-150" onClick={fixRotation}>Fix Rotation 270</button>
                        </div>
                    </div>
                    <button
                        disabled={isLoaded}
                        className="px-3 py-2 bg-cyan-700 text-white mt-3 rounded-3xl hover:bg-cyan-800 cursor-pointer duration-150 disabled:opacity-50 w-full"
                    >
                        {isLoaded ? <>
                            {percentUpload !== 100 ? `Uploading ${percentUpload}%` : 'Generating...'}
                        </> : 'Generate'}
                    </button>
                </form>
                {imageData && <img src={imageData} alt="" className="border mt-3" />}
            </div>
        </div>
    )
}